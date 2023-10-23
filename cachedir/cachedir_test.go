package cachedir

import (
	"os"
	"testing"
)

const (
	testName = "goasn-golang-test"
)

func TestMakeCacheDir(t *testing.T) {
	cd, err := GetCacheDir(testName)
	if err != nil {
		t.Fatal(err)
	}

	err = os.RemoveAll(cd)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ {
		_, err := MakeCacheDir(testName)
		if err != nil {
			t.Fatal(err)
		}
	}

	err = os.RemoveAll(cd)
	if err != nil {
		t.Fatal(err)
	}
}
