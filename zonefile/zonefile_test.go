package zonefile

import (
	"bytes"
	"io"
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
	expectedZoneV4 = []byte(zoneHeader + `8.0.0.0/8 12|8.0.0.0/8|US|arin|
9.0.0.0/8 13|9.0.0.0/8|TW|apnic|
`)
	expectedZoneV6 = []byte(zoneHeader + `2404:c800::/32 22|2404:c800::/32|US|arin|
2505:c900::/32 23|2505:c900::/32|TW|apnic|
`)
)

func checkFileContents(zonefile string, expected []byte) (bool, error) {
	f, err := os.Open(zonefile)
	if err != nil {
		return false, err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return false, err
	}

	return bytes.Equal(b, expected), nil
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

	var ok bool
	ok, err = checkFileContents(zoneV4Full, expectedZoneV4)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("V4 zone was not as expected")
	}

	ok, err = checkFileContents(zoneV6Full, expectedZoneV6)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("V6 zone was not as expected")
	}

	err = os.RemoveAll(tempDir)
	if err != nil {
		t.Fatal(err)
	}
}
