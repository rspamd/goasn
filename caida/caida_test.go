package caida

import (
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/rspamd/goasn/ir"
	"github.com/rspamd/goasn/log"

	"github.com/asergeyev/nradix"
)

func stubIANA(uint32) ir.IRID { return ir.ARIN }

func writeGzip(t *testing.T, path, body string) {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLatestFromLog(t *testing.T) {
	log.SetupLogger(false)
	dir := t.TempDir()
	logPath := filepath.Join(dir, "pfx2as-creation.log")
	body := "# header\n" +
		"100\t1700000000\t2023/11/routeviews-rv2-20231114-1200.pfx2as.gz\n" +
		"101\t1700010000\t2023/11/routeviews-rv2-20231114-1500.pfx2as.gz\n" +
		"99\t1699990000\t2023/11/routeviews-rv2-20231114-0900.pfx2as.gz\n"
	if err := os.WriteFile(logPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := latestFromLog(logPath,
		"https://publicdata.caida.org/datasets/routing/routeviews-prefix2as/pfx2as-creation.log")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://publicdata.caida.org/datasets/routing/routeviews-prefix2as/2023/11/routeviews-rv2-20231114-1500.pfx2as.gz"
	if got != want {
		t.Fatalf("latestFromLog: got %q want %q", got, want)
	}
}

func TestParsePfx2AS(t *testing.T) {
	log.SetupLogger(false)
	dir := t.TempDir()
	data := "" +
		"1.0.0.0\t24\t13335\n" +
		"8.8.8.0\t24\t15169\n" +
		// multi-origin (underscore) -> first wins
		"192.0.2.0\t24\t64500_64501\n" +
		// AS-set (comma) -> first wins
		"198.51.100.0\t24\t64502,64503\n" +
		// junk / malformed lines are dropped
		"# comment line\n" +
		"bogus line\n" +
		"\n"
	gzPath := filepath.Join(dir, "routeviews-rv2-20231114-1500.pfx2as.gz")
	writeGzip(t, gzPath, data)

	got, err := parsePfx2AS(gzPath, stubIANA, nradix.NewTree(0))
	if err != nil {
		t.Fatal(err)
	}
	expect := map[string]uint32{
		"1.0.0.0/24":     13335,
		"8.8.8.0/24":     15169,
		"192.0.2.0/24":   64500,
		"198.51.100.0/24": 64502,
	}
	if len(got) != len(expect) {
		t.Fatalf("prefix count: got %d want %d (%v)", len(got), len(expect), got)
	}
	for k, v := range expect {
		if got[k] != v {
			t.Errorf("%s: got %d want %d", k, got[k], v)
		}
	}
}

func TestParsePfx2ASReservedFilter(t *testing.T) {
	log.SetupLogger(false)
	dir := t.TempDir()
	data := "" +
		"10.0.0.0\t8\t65000\n" + // private, should be filtered
		"1.0.0.0\t24\t13335\n"
	gzPath := filepath.Join(dir, "t.pfx2as.gz")
	writeGzip(t, gzPath, data)

	reserved := nradix.NewTree(0)
	if err := reserved.AddCIDR("10.0.0.0/8", struct{}{}); err != nil {
		t.Fatal(err)
	}

	got, err := parsePfx2AS(gzPath, stubIANA, reserved)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got["10.0.0.0/8"]; ok {
		t.Fatal("reserved 10.0.0.0/8 should have been filtered")
	}
	if got["1.0.0.0/24"] != 13335 {
		t.Fatalf("1.0.0.0/24: got %d want 13335", got["1.0.0.0/24"])
	}
}

func TestParsePfx2ASBogusASNFilter(t *testing.T) {
	log.SetupLogger(false)
	dir := t.TempDir()
	data := "" +
		"1.0.0.0\t24\t13335\n" +
		"2.0.0.0\t24\t99999\n"
	gzPath := filepath.Join(dir, "t.pfx2as.gz")
	writeGzip(t, gzPath, data)

	ianaStub := func(asn uint32) ir.IRID {
		if asn == 99999 {
			return ir.RESERVED
		}
		return ir.ARIN
	}
	got, err := parsePfx2AS(gzPath, ianaStub, nradix.NewTree(0))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got["2.0.0.0/24"]; ok {
		t.Fatal("ASN flagged RESERVED should have been filtered")
	}
	if got["1.0.0.0/24"] != 13335 {
		t.Fatalf("1.0.0.0/24: got %d want 13335", got["1.0.0.0/24"])
	}
}
