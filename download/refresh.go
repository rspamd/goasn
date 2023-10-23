package download

import (
	"sync"

	"go.uber.org/zap"

	"github.com/rspamd/goasn/log"
)

func RefreshSources(appCacheDir string, sources []string) bool {
	var wg sync.WaitGroup
	allGood := true

	for _, url := range sources {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			err := DownloadSource(appCacheDir, url)
			if err != nil {
				log.Logger.Error("failed to get update",
					zap.String("url", url), zap.Error(err))
				allGood = false
			}
		}(url)
	}
	wg.Wait()
	return allGood
}
