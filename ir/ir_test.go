package ir

import (
	"os"
	"path"
	"runtime"
	"testing"
)

var (
	expectedValues = map[uint32]IRASNInfo{
		uint32(80): IRASNInfo{
			IR:      ARIN,
			Country: "US",
		},
		uint32(1768): IRASNInfo{
			IR:      APNIC,
			Country: "TW",
		},
	}
)

func TestReadIRData(t *testing.T) {
	_, ourFile, _, _ := runtime.Caller(0)
	testDataDir := path.Join(path.Dir(ourFile), "testdata")

	testFileInfo, err := os.ReadDir(testDataDir)
	if err != nil {
		t.Fatal(err)
	}
	testFileNames := make([]string, len(testFileInfo))
	for i := 0; i < len(testFileInfo); i++ {
		testFileNames[i] = testFileInfo[i].Name()
	}

	irMap, err := ReadIRData(testDataDir, testFileNames)
	if err != nil {
		t.Fatal(err)
	}

	for k, v := range expectedValues {
		realVal := irMap[k]
		if realVal.Country != v.Country {
			t.Fatalf("%d: expected country %s got %s", k, v.Country, realVal.Country)
		}
		if realVal.IR != v.IR {
			t.Fatalf("%d: expected IR %s got %s", k, v.IR, realVal.IR)
		}
	}
}
