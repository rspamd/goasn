package iana

import (
	"path"
	"runtime"
	"testing"

	"github.com/rspamd/goasn/log"
)

var (
	expectedEntries = []string{
		"127.0.0.1/32",
		"127.0.0.0/24",
		"10.1.0.0/16",
		"10.0.0.0/8",
		"255.255.255.255/32",
	}
	unexpectedEntries = []string{
		"1.1.1.1/32",
		"8.8.8.0/24",
		"223.0.0.1/24",
	}
)

func TestIANAIP4(t *testing.T) {
	log.SetupLogger(false)

	_, ourFile, _, _ := runtime.Caller(0)
	testDataDir := path.Join(path.Dir(ourFile), "testdata")

	tree, err := GetReservedIP4(testDataDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, e := range expectedEntries {
		inf, err := tree.FindCIDR(e)
		if err != nil {
			t.Fatal(err)
		}
		if inf == nil {
			t.Fatalf("didn't find expected entry %s", e)
		}
	}

	for _, e := range unexpectedEntries {
		inf, err := tree.FindCIDR(e)
		if err != nil {
			t.Fatal(err)
		}
		if inf != nil {
			t.Fatalf("found unexpected entry %s", e)
		}
	}

}
