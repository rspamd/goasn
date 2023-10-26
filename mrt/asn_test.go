package mrt

import (
	"path"
	"runtime"
	"testing"

	"github.com/rspamd/goasn/ir"
	"github.com/rspamd/goasn/log"

	"github.com/asergeyev/nradix"
)

func stubIANA(asnNo uint32) ir.IRID {
	return ir.ARIN
}

func TestASNFromBGP(t *testing.T) {
	log.SetupLogger(false)
	_, ourFile, _, _ := runtime.Caller(0)
	testDataDir := path.Join(path.Dir(ourFile), "testdata")

	fake4 := nradix.NewTree(0)
	fake6 := nradix.NewTree(0)

	bgpInfo := ASNFromBGP(testDataDir, stubIANA, "", fake4, fake6)
	if bgpInfo.Err != nil {
		t.Fatal(bgpInfo.Err)
	}
}
