package download

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/rspamd/goasn/log"
)

func TestRefresh(t *testing.T) {
	log.SetupLogger(false)

	_, ourFile, _, _ := runtime.Caller(0)
	testDataDir := path.Join(path.Dir(ourFile), "testdata")

	ts := httptest.NewServer(http.FileServer(http.Dir(testDataDir)))
	defer ts.Close()

	tempDir, err := os.MkdirTemp("", "goasn-test")
	if err != nil {
		t.Fatal(err)
	}

	urlList := make([]string, 0)
	for _, v := range []string{"file1", "file2"} {
		urlList = append(urlList, ts.URL+"/"+v)
	}

	for i := 0; i < 2; i++ {
		result := RefreshSources(tempDir, urlList)
		if result.AnyError {
			t.Fatal("sources failed to refresh: error occurred")
		}
		if !result.AnyUpdated {
			t.Fatal("sources failed to refresh: no updates")
		}
	}

	err = os.RemoveAll(tempDir)
	if err != nil {
		t.Fatal(err)
	}
}
