package sources

import (
	"net/url"
	"path"
)

func Basename(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	return path.Base(u.Path), nil
}

func Basenames(urlList []string) ([]string, error) {

	baseNames := make([]string, len(urlList))

	for i, resourceURL := range urlList {
		u, err := url.Parse(resourceURL)
		if err != nil {
			return baseNames, err
		}
		baseNames[i] = path.Base(u.Path)
	}

	return baseNames, nil
}
