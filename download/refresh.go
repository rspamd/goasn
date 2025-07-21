package download

import (
	"sync"

	"go.uber.org/zap"

	"github.com/rspamd/goasn/log"
)

type RefreshResult struct {
	AnyUpdated   bool
	AnyError     bool
	ErrorCount   int
	UpdatedCount int
}

func RefreshSources(appCacheDir string, sources []string) RefreshResult {
	var wg sync.WaitGroup
	var mu sync.Mutex
	result := RefreshResult{}

	for _, url := range sources {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			downloaded, err := DownloadSource(appCacheDir, url)
			mu.Lock()
			if err != nil {
				log.Logger.Error("failed to get update",
					zap.String("url", url), zap.Error(err))
				result.AnyError = true
				result.ErrorCount++
			} else if downloaded {
				result.AnyUpdated = true
				result.UpdatedCount++
			}
			mu.Unlock()
		}(url)
	}
	wg.Wait()
	return result
}
