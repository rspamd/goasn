package cachedir

import (
	"fmt"
	"os"
	"path"
)

func GetCacheDir(appName string) (string, error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return path.Join(userCacheDir, appName), nil
}

func MakeCacheDir(appName string) (string, error) {
	ourDir, err := GetCacheDir(appName)
	if err != nil {
		return "", fmt.Errorf("failed to get user cache dir: %v", err)
	}

	fi, err := os.Stat(ourDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(ourDir, 0750)
			if err != nil {
				return ourDir, fmt.Errorf("failed to create our cache dir(%s): %v", ourDir, err)
			}
		} else {
			return ourDir, fmt.Errorf("unexpected error stat'ing our cache dir(%s): %v", ourDir, err)
		}
	} else {
		if !fi.Mode().IsDir() {
			return ourDir, fmt.Errorf("our cache dir(%s) is not a directory", ourDir)
		}
	}
	return ourDir, nil
}
