package iana

import (
	"path"
	"runtime"
	"testing"

	"github.com/rspamd/goasn/log"
)

var (
	expectedEntries6 = []string{
		"100::1/128",
		"600:803:29c::/48", // 0400::/6
		"fe80::943d:77d1:97a4:dacc/128",
	}
	unexpectedEntries6 = []string{
		"2001:1af8:4700::1/128",
	}
)

func TestIANAIP6(t *testing.T) {
	log.SetupLogger(false)

	_, ourFile, _, _ := runtime.Caller(0)
	testDataDir := path.Join(path.Dir(ourFile), "testdata")

	tree, err := GetReservedIP6(testDataDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, e := range expectedEntries6 {
		inf, err := tree.FindCIDR(e)
		if err != nil {
			t.Fatal(err)
		}
		if inf == nil {
			t.Fatalf("didn't find expected entry %s", e)
		}
	}

	for _, e := range unexpectedEntries6 {
		inf, err := tree.FindCIDR(e)
		if err != nil {
			t.Fatal(err)
		}
		if inf != nil {
			t.Fatalf("found unexpected entry %s", e)
		}
	}

}
