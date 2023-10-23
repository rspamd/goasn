package iana

import (
	"path"
	"runtime"
	"testing"

	"github.com/rspamd/goasn/ir"
)

var (
	expectedMapping = map[uint32]ir.IRID{
		uint32(0):      ir.RESERVED,
		uint32(1879):   ir.RIPE,
		uint32(2048):   ir.ARIN,
		uint32(64513):  ir.RESERVED,
		uint32(262144): ir.LACNIC,
		uint32(329727): ir.AFRINIC,
		uint32(153914): ir.UNALLOCATED,
	}
)

func TestIANAASN(t *testing.T) {
	_, ourFile, _, _ := runtime.Caller(0)
	testDataDir := path.Join(path.Dir(ourFile), "testdata")

	asnAlloc, err := ReadIANAASN(testDataDir)
	if err != nil {
		t.Fatal(err)
	}

	for k, v := range expectedMapping {
		alloc := asnAlloc(k)
		if alloc != v {
			t.Fatalf("expected <%s> for ASN %d got <%s>", v, k, alloc)
		}
	}
}
