// Package caida parses CAIDA's RouteViews "pfx2as" daily dumps and exposes them
// as prefix→origin-AS maps that can be merged into the BGP-derived data.
//
// The CAIDA datasets aggregate several RouteViews collectors, which gives much
// broader peer coverage than a single RIPE RIS collector. We use them as an
// additive enrichment: prefixes already present in the BGP view keep their
// origin AS; anything new from CAIDA is merged in.
//
// Dataset layout:
//
//	<base>/pfx2as-creation.log
//	<base>/YYYY/MM/routeviews-*-YYYYMMDD-HHMM.pfx2as.gz
//
// The "creation log" has one entry per published file:
//
//	<serial>\t<epoch>\t<relative_path>
//
// We pick the entry with the highest epoch and fetch that file.
package caida

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"maps"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rspamd/goasn/download"
	"github.com/rspamd/goasn/ir"
	"github.com/rspamd/goasn/log"
	"github.com/rspamd/goasn/sources"

	"github.com/asergeyev/nradix"
	"go.uber.org/zap"
)

type PfxResult struct {
	V4 map[string]uint32
	V6 map[string]uint32
}

func NewPfxResult() *PfxResult {
	return &PfxResult{
		V4: make(map[string]uint32),
		V6: make(map[string]uint32),
	}
}

// RefreshResult summarises a Refresh() call.
type RefreshResult struct {
	AnyUpdated   bool
	ErrorCount   int
	UpdatedCount int
}

// datasetDir returns (and creates if missing) the per-dataset cache
// subdirectory for a given CAIDA log URL. The v4 and v6 datasets both publish
// pfx2as-creation.log, so we can't share a single directory.
func datasetDir(appCacheDir, logURL string) (string, error) {
	dir := filepath.Join(appCacheDir, sources.CAIDADatasetDir(logURL))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func logLocalPath(datasetDir, logURL string) (string, error) {
	u, err := url.Parse(logURL)
	if err != nil {
		return "", err
	}
	return filepath.Join(datasetDir, path.Base(u.Path)), nil
}

// latestFromLog scans a CAIDA pfx2as-creation.log and returns the absolute URL
// of the most recent .pfx2as.gz file referenced in it.
func latestFromLog(logPath string, logURL string) (string, error) {
	f, err := os.Open(logPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	u, err := url.Parse(logURL)
	if err != nil {
		return "", err
	}
	baseDir := path.Dir(u.Path)

	var latestEpoch int64 = -1
	var latestRel string
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		epoch, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			continue
		}
		if epoch > latestEpoch {
			latestEpoch = epoch
			latestRel = fields[2]
		}
	}
	if err := s.Err(); err != nil {
		return "", fmt.Errorf("scan %s: %v", logPath, err)
	}
	if latestRel == "" {
		return "", fmt.Errorf("no entries found in CAIDA log %s", logPath)
	}

	latestURL := *u
	latestURL.Path = path.Join(baseDir, latestRel)
	return latestURL.String(), nil
}

// parsePfx2AS reads a gzipped CAIDA pfx2as file and returns a prefix→origin-AS
// map. Prefixes whose origin AS is bogus or whose prefix falls in an IANA
// reserved range (when a filter is provided) are skipped, mirroring the BGP
// path in mrt/asn.go.
//
// CAIDA origin fields can be:
//
//	12345              - single ASN
//	12345_67890        - multi-origin set (underscore)
//	12345,67890        - AS set (comma)
//
// We take the first numeric ASN; exotic cases are dropped.
func parsePfx2AS(filePath string, ianaASN func(uint32) ir.IRID, reserved *nradix.Tree) (map[string]uint32, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("gzip open %s: %v", filePath, err)
	}
	defer gr.Close()

	out := make(map[string]uint32)
	skippedBogus := 0
	skippedReserved := 0
	s := bufio.NewScanner(gr)
	s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for s.Scan() {
		line := s.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 3 {
			continue
		}
		prefix := fields[0] + "/" + fields[1]
		asField := fields[2]
		for _, sep := range []string{",", "_"} {
			if idx := strings.Index(asField, sep); idx >= 0 {
				asField = asField[:idx]
			}
		}
		asn, err := strconv.ParseUint(asField, 10, 32)
		if err != nil {
			continue
		}
		if ianaASN != nil {
			id := ianaASN(uint32(asn))
			if id == ir.UNKNOWN || id == ir.RESERVED {
				skippedBogus++
				continue
			}
		}
		if reserved != nil {
			inf, _ := reserved.FindCIDR(prefix)
			if inf != nil {
				skippedReserved++
				continue
			}
		}
		out[prefix] = uint32(asn)
	}
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %v", filePath, err)
	}
	log.Logger.Debug("parsed CAIDA pfx2as",
		zap.String("file", filePath),
		zap.Int("prefixes", len(out)),
		zap.Int("skipped_bogus_asn", skippedBogus),
		zap.Int("skipped_reserved_prefix", skippedReserved))
	return out, nil
}

// Refresh downloads each CAIDA pfx2as-creation.log and then the newest file it
// references. Updates are reported via RefreshResult so callers can skip
// regeneration when nothing changed.
func Refresh(appCacheDir string) RefreshResult {
	var result RefreshResult
	for _, logURL := range sources.GetCAIDASources() {
		dir, err := datasetDir(appCacheDir, logURL)
		if err != nil {
			log.Logger.Error("failed to prepare CAIDA dataset dir",
				zap.String("url", logURL), zap.Error(err))
			result.ErrorCount++
			continue
		}

		logUpdated, err := download.DownloadSource(dir, logURL)
		if err != nil {
			log.Logger.Error("failed to get CAIDA log update",
				zap.String("url", logURL), zap.Error(err))
			result.ErrorCount++
			continue
		}
		if logUpdated {
			result.UpdatedCount++
			result.AnyUpdated = true
		}

		logPath, err := logLocalPath(dir, logURL)
		if err != nil {
			log.Logger.Error("bad CAIDA log URL",
				zap.String("url", logURL), zap.Error(err))
			result.ErrorCount++
			continue
		}
		latestURL, err := latestFromLog(logPath, logURL)
		if err != nil {
			log.Logger.Error("failed to parse CAIDA log",
				zap.String("path", logPath), zap.Error(err))
			result.ErrorCount++
			continue
		}
		fileUpdated, err := download.DownloadSource(dir, latestURL)
		if err != nil {
			log.Logger.Error("failed to download CAIDA pfx2as file",
				zap.String("url", latestURL), zap.Error(err))
			result.ErrorCount++
			continue
		}
		if fileUpdated {
			result.UpdatedCount++
			result.AnyUpdated = true
		}
		log.Logger.Debug("CAIDA refresh",
			zap.String("log", logURL),
			zap.String("latest", latestURL),
			zap.Bool("log_updated", logUpdated),
			zap.Bool("file_updated", fileUpdated))
	}
	return result
}

// Load returns prefix→AS maps from the most recent cached CAIDA files. It does
// not perform any network I/O; call Refresh first if you want fresh data.
func Load(appCacheDir string, ianaASN func(uint32) ir.IRID, reserved4, reserved6 *nradix.Tree) (*PfxResult, error) {
	res := NewPfxResult()
	for _, logURL := range sources.GetCAIDASources() {
		dir, err := datasetDir(appCacheDir, logURL)
		if err != nil {
			return res, err
		}
		logPath, err := logLocalPath(dir, logURL)
		if err != nil {
			return res, err
		}
		latestURL, err := latestFromLog(logPath, logURL)
		if err != nil {
			return res, err
		}
		u, err := url.Parse(latestURL)
		if err != nil {
			return res, err
		}
		dataPath := filepath.Join(dir, path.Base(u.Path))

		isV6 := logURL == sources.CAIDA_RV6_PFX2AS_LOG
		var target map[string]uint32
		var reserved *nradix.Tree
		if isV6 {
			target = res.V6
			reserved = reserved6
		} else {
			target = res.V4
			reserved = reserved4
		}
		m, err := parsePfx2AS(dataPath, ianaASN, reserved)
		if err != nil {
			return res, err
		}
		maps.Copy(target, m)
	}
	return res, nil
}
