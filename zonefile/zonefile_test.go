package zonefile

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/rspamd/goasn/ir"
	"github.com/rspamd/goasn/log"
)

var (
	fakeV4 = map[string]uint32{
		"8.0.0.0/8": uint32(12),
		"9.0.0.0/8": uint32(13),
	}
	fakeV6 = map[string]uint32{
		"2404:c800::/32": uint32(22),
		"2505:c900::/32": uint32(23),
	}
	fakeASN = map[uint32]ir.IRASNInfo{
		uint32(12): ir.IRASNInfo{
			IR:      ir.ARIN,
			Country: "US",
		},
		uint32(13): ir.IRASNInfo{
			IR:      ir.APNIC,
			Country: "TW",
		},
		uint32(22): ir.IRASNInfo{
			IR:      ir.ARIN,
			Country: "US",
		},
		uint32(23): ir.IRASNInfo{
			IR:      ir.APNIC,
			Country: "TW",
		},
	}
	expectedZoneV4Entries = map[string]struct{}{
		"8.0.0.0/8 12|8.0.0.0/8|US|arin|":  struct{}{},
		"9.0.0.0/8 13|9.0.0.0/8|TW|apnic|": struct{}{},
	}

	expectedZoneV6Entries = map[string]struct{}{
		"2404:c800::/32 22|2404:c800::/32|US|arin|":  struct{}{},
		"2505:c900::/32 23|2505:c900::/32|TW|apnic|": struct{}{},
	}
)

func checkFileContents(zonefile string, expected map[string]struct{}) error {
	f, err := os.Open(zonefile)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(f)
	for i := 0; i < 2; i++ {
		scanner.Scan() // skip header
	}
	for scanner.Scan() {
		txt := scanner.Text()
		_, ok := expected[txt]
		if !ok {
			return fmt.Errorf("found unexpected entry: %s", txt)
		} else {
			delete(expected, txt)
		}
	}

	err = scanner.Err()
	if err != nil {
		return err
	}

	return f.Close()
}

func TestGenerateZones(t *testing.T) {
	log.SetupLogger(false)

	tempDir, err := os.MkdirTemp("", "goasn-test")
	if err != nil {
		t.Fatal(err)
	}

	zoneV4Full := path.Join(tempDir, "goasn4.zone")
	zoneV6Full := path.Join(tempDir, "goasn6.zone")

	err = GenerateZones(fakeASN, fakeV4, zoneV4Full, fakeV6, zoneV6Full)
	if err != nil {
		t.Fatal(err)
	}

	err = checkFileContents(zoneV4Full, expectedZoneV4Entries)
	if err != nil {
		t.Fatal(err)
	}

	err = checkFileContents(zoneV6Full, expectedZoneV6Entries)
	if err != nil {
		t.Fatal(err)
	}

	err = os.RemoveAll(tempDir)
	if err != nil {
		t.Fatal(err)
	}
}
