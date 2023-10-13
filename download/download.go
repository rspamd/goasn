package download

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"go.uber.org/zap"

	"github.com/rspamd/goasn/log"
)

var (
	httpClient = &http.Client{}
)

func lastModFromHeaders(hdrs http.Header, resourceURL string) (t time.Time, err error) {
	lastMod := hdrs.Get("Last-Modified")
	if lastMod == "" {
		return t, fmt.Errorf("no last modified time for URL: %s", resourceURL)
	}
	t, err = time.Parse(time.RFC1123, lastMod)
	if err != nil {
		return t, fmt.Errorf("couldn't parse last-modified time(%s) for URL(%s): %v", lastMod, resourceURL, err)
	}
	return t, nil
}

func CheckUpdate(resourceURL string, fileModTime time.Time) (bool, error) {
	req, err := http.NewRequest("HEAD", resourceURL, nil)
	if err != nil {
		return false, fmt.Errorf("failed to prepare HEAD request to %s: %v", resourceURL, err)
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("HEAD request to %s failed: %v", resourceURL, err)
	}
	if res.StatusCode != 200 {
		return false, fmt.Errorf("HEAD request to %s returned bad status: %d", resourceURL, res.StatusCode)
	}
	t, err := lastModFromHeaders(res.Header, resourceURL)
	if err != nil {
		return false, err
	}
	if !t.After(fileModTime) {
		return false, nil
	}
	return true, nil
}

func DownloadSource(ourDir string, resourceURL string) error {
	wantDownload := true

	u, err := url.Parse(resourceURL)
	if err != nil {
		return fmt.Errorf("couldn't parse resource URL(%s): %v", resourceURL, err)
	}
	fName := path.Base(u.Path)

	fPath := path.Join(ourDir, fName)
	fi, err := os.Stat(fPath)
	if err != nil {
		if os.IsNotExist(err) {
			wantDownload = true
		} else {
			return fmt.Errorf("unexpected error stat'ing file(%s): %v", fPath, err)
		}
	} else {
		wantDownload, err = CheckUpdate(resourceURL, fi.ModTime())
		if err != nil {
			return fmt.Errorf("checking for update(%s) failed: %v", resourceURL, err)
		}
	}

	if !wantDownload {
		log.Logger.Debug("no update needed", zap.String("url", resourceURL))
		return nil
	}

	req, err := http.NewRequest("GET", resourceURL, nil)
	if err != nil {
		return fmt.Errorf("failed to prepare GET request to %s: %v", resourceURL, err)
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("GET request to %s failed: %v", resourceURL, err)
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("GET request to %s returned bad status: %d", resourceURL, res.StatusCode)
	}
	defer res.Body.Close()

	t, err := lastModFromHeaders(res.Header, resourceURL)
	if err != nil {
		return err
	}

	swapPath := fPath + ".swp"
	f, err := os.Create(swapPath)
	if err != nil {
		return fmt.Errorf("failed to create file(%s): %v", swapPath, err)
	}

	_, err = io.Copy(f, res.Body)

	if err != nil {
		closeError := f.Close()
		if closeError != nil {
			closeError = fmt.Errorf("failed to close file(%s): %v", swapPath, closeError)
		}
		return errors.Join(closeError, fmt.Errorf("copy error: %v", err))
	}

	err = f.Close()
	if err != nil {
		return fmt.Errorf("close error: %v", err)
	}

	err = os.Chtimes(swapPath, time.Time{}, t)
	if err != nil {
		return fmt.Errorf("chtimes error: %v", err)
	}

	err = os.Rename(swapPath, fPath)
	if err != nil {
		return fmt.Errorf("failed to rename %s to %s: %v", swapPath, fPath, err)
	}
	return nil
}
