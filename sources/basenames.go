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

func MustBasename(urlStr string) string {
	res, err := Basename(urlStr)
	if err != nil {
		panic(err)
	}
	return res
}

func Basenames(urlList []string) ([]string, error) {

	baseNames := make([]string, len(urlList))

	for i, resourceURL := range urlList {
		b, err := Basename(resourceURL)
		if err != nil {
			return baseNames, err
		}
		baseNames[i] = b
	}

	return baseNames, nil
}

func MustBasenames(urlList []string) []string {
	res, err := Basenames(urlList)
	if err != nil {
		panic(err)
	}
	return res
}
