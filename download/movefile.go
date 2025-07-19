package download

import (
	"os"
)

// MoveFile atomically moves src to dst, overwriting dst if it exists.
func MoveFile(src, dst string) error {
	return os.Rename(src, dst)
}
