package mrt

import (
	"path"
	"runtime"
	"testing"

	"github.com/rspamd/goasn/ir"
	"github.com/rspamd/goasn/log"
)

func stubIANA(asnNo uint32) ir.IRID {
	return ir.ARIN
}

func TestASNFromBGP(t *testing.T) {
	log.SetupLogger(false)
	_, ourFile, _, _ := runtime.Caller(0)
	testDataDir := path.Join(path.Dir(ourFile), "testdata")

	bgpInfo := ASNFromBGP(testDataDir, stubIANA, "")
	if bgpInfo.Err != nil {
		t.Fatal(bgpInfo.Err)
	}
}
