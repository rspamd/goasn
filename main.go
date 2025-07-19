package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rspamd/goasn/cachedir"
	"github.com/rspamd/goasn/download"
	"github.com/rspamd/goasn/iana"
	"github.com/rspamd/goasn/ir"
	"github.com/rspamd/goasn/log"
	"github.com/rspamd/goasn/mrt"
	"github.com/rspamd/goasn/sources"
	"github.com/rspamd/goasn/zonefile"

	"github.com/asergeyev/nradix"
	flag "github.com/spf13/pflag"
	"go.uber.org/zap"
)

const (
	APP_NAME = "goasn"
)

var (
	debug        bool
	downloadASN  bool
	downloadBGP  bool
	onUpdateOnly bool
	rejectFile   string
	zoneV4       string
	zoneV6       string
	cacheDir     string
	zoneTmpExt   string
)

func main() {
	var appCacheDir string
	var err error
	log.Logger.Info("goasn application started")
	if cacheDir != "" {
		// Create the directory if it doesn't exist
		err = os.MkdirAll(cacheDir, 0o755)
		if err != nil {
			log.Logger.Fatal("failed to create cache directory",
				zap.Error(err))
		}
		appCacheDir = cacheDir
	} else {
		appCacheDir, err = cachedir.MakeCacheDir(APP_NAME)
		if err != nil {
			log.Logger.Fatal("failed to create cache directory",
				zap.Error(err))
		}
	}

	// Ensure zone file directories exist and are writable
	for _, zonePath := range []string{zoneV4, zoneV6} {
		if zonePath == "" {
			continue
		}
		dir := filepath.Dir(zonePath)
		err := os.MkdirAll(dir, 0o755)
		if err != nil {
			log.Logger.Fatal("failed to create zone file directory",
				zap.String("dir", dir), zap.Error(err))
		}
		tmpFile := zonePath + zoneTmpExt
		f, err := os.Create(tmpFile)
		if err != nil {
			log.Logger.Fatal("failed to create temp zone file",
				zap.String("file", tmpFile), zap.Error(err))
		}
		_, err = f.Write([]byte("test"))
		if err != nil {
			f.Close()
			os.Remove(tmpFile)
			log.Logger.Fatal("failed to write to temp zone file",
				zap.String("file", tmpFile), zap.Error(err))
		}
		f.Close()
		os.Remove(tmpFile)
	}

	toRefresh := make([]string, 0)
	if downloadASN {
		toRefresh = append(toRefresh, sources.GetASNSources()...)
	}
	if downloadBGP {
		toRefresh = append(toRefresh, sources.BGP_LATEST)
	}

	result := download.RefreshSources(appCacheDir, toRefresh)
	if result.AnyError {
		log.Logger.Warn("failed to download sources", zap.Int("error_count", result.ErrorCount))
	} else if !result.AnyUpdated {
		log.Logger.Info("all sources up to date")
	}

	if result.AnyUpdated {
		log.Logger.Info("sources succesfully updated", zap.Int("updated_count", result.UpdatedCount))
	}

	if (zoneV4 != "" || zoneV6 != "") && onUpdateOnly && !result.AnyUpdated {
		skip := true
		for _, zonePath := range []string{zoneV4, zoneV6} {
			if zonePath == "" {
				continue
			}
			f, err := os.Open(zonePath)
			if err != nil {
				skip = false
				break
			}
			buf := make([]byte, 4096)
			n, _ := f.Read(buf)
			f.Close()
			content := string(buf[:n])
			lines := strings.Split(content, "\n")
			if len(lines) < 2 ||
				!strings.HasPrefix(lines[0], "$SOA ") ||
				!strings.HasPrefix(lines[1], "$NS ") {
				skip = false
				break
			}
		}
		if skip {
			log.Logger.Info("skipping zone file generation")
			return
		}
	}

	log.Logger.Info("starting zone file generation")
	IRDataFiles := sources.MustBasenames(sources.GetRIRASN())
	asnToIRInfo, err := ir.ReadIRData(appCacheDir, IRDataFiles)
	if err != nil {
		log.Logger.Fatal("failed to read ASN info", zap.Error(err))
	}

	ianaASN, err := iana.ReadIANAASN(appCacheDir)
	if err != nil {
		log.Logger.Fatal("failed to read IANA ASN info", zap.Error(err))
	}

	var reservedV4 *nradix.Tree
	var reservedV6 *nradix.Tree

	if zoneV4 != "" {
		var err error
		reservedV4, err = iana.GetReservedIP4(appCacheDir)
		if err != nil {
			log.Logger.Fatal("failed to read IANA IP4 info", zap.Error(err))
		}
	} else {
		reservedV4 = nradix.NewTree(0)
	}

	if zoneV6 != "" {
		var err error
		reservedV6, err = iana.GetReservedIP6(appCacheDir)
		if err != nil {
			log.Logger.Fatal("failed to read IANA IP6 info", zap.Error(err))
		}
	} else {
		reservedV6 = nradix.NewTree(0)
	}

	bgpInfo := mrt.ASNFromBGP(appCacheDir, ianaASN, rejectFile, reservedV4, reservedV6)
	if bgpInfo.Err != nil {
		log.Logger.Fatal("failed to process MRT", zap.Error(bgpInfo.Err))
	}
	if bgpInfo.ParseErrorCount > 0 {
		log.Logger.Error("MRT parsing errors occurred",
			zap.Int("count", bgpInfo.ParseErrorCount),
			zap.Any("errors", bgpInfo.ParseErrors))
	}

	// Write zone files to same dir as destination, with temp extension
	var tmpZoneV4, tmpZoneV6 string
	if zoneV4 != "" {
		tmpZoneV4 = zoneV4 + zoneTmpExt
	} else {
		tmpZoneV4 = ""
	}
	if zoneV6 != "" {
		tmpZoneV6 = zoneV6 + zoneTmpExt
	} else {
		tmpZoneV6 = ""
	}

	err = zonefile.GenerateZones(asnToIRInfo, bgpInfo.V4, tmpZoneV4, bgpInfo.V6, tmpZoneV6)
	if err != nil {
		log.Logger.Fatal("failed to generate zone", zap.Error(err))
	}

	// Atomically move files from temp to destination
	if zoneV4 != "" {
		err = download.MoveFile(tmpZoneV4, zoneV4)
		if err != nil {
			log.Logger.Fatal("failed to move V4 zone file", zap.Error(err))
		} else {
			log.Logger.Info("finished writing V4 zone file", zap.String("file", zoneV4))
		}
	}
	if zoneV6 != "" {
		err = download.MoveFile(tmpZoneV6, zoneV6)
		if err != nil {
			log.Logger.Fatal("failed to move V6 zone file", zap.Error(err))
		} else {
			log.Logger.Info("finished writing V6 zone file", zap.String("file", zoneV6))
		}
	}
}

func init() {
	flag.BoolVar(&debug, "debug", false, "enable debug logging")
	flag.BoolVar(&downloadASN, "download-asn", false, "download RIR data")
	flag.BoolVar(&downloadBGP, "download-bgp", false, "download MRT data")
	flag.BoolVar(&onUpdateOnly, "on-update-only", false, "generate zones only if resources were updated")
	flag.StringVar(&rejectFile, "reject", "", "path to write unparseable entries to")
	flag.StringVar(&zoneV4, "file-v4", "", "path to V4 zonefile")
	flag.StringVar(&zoneV6, "file-v6", "", "path to V6 zonefile")
	flag.StringVar(&cacheDir, "cache-dir", "", "directory for cache files")
	flag.StringVar(&zoneTmpExt, "zone-tmp-ext", ".tmp", "temp extension for zone files, default to: .tmp")
	flag.Parse()

	err := log.SetupLogger(debug)
	if err != nil {
		panic(err)
	}
}
