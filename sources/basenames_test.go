package sources

import (
	"reflect"
	"testing"
)

var (
	expectedBasenames = []string{
		"delegated-afrinic-latest",
		"delegated-apnic-latest",
		"delegated-arin-extended-latest",
		"delegated-lacnic-latest",
		"delegated-ripencc-latest",
		"latest-bview.gz",
		"as-numbers.xml",
		"ipv4-address-space.xml",
		"ipv6-address-space.xml",
	}
)

func TestBasenames(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatal(r)
		}
	}()

	baseNames := MustBasenames(GetAllSources())
	if !reflect.DeepEqual(baseNames, expectedBasenames) {
		t.Fatalf("%s (actual) != %s (expected)", baseNames, expectedBasenames)
	}
}

func TestBasename(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatal(r)
		}
	}()

	baseName := MustBasename("http://example.net/foo.txt")
	if baseName != "foo.txt" {
		t.Fatalf("expected foo.txt got %s", baseName)
	}
}
