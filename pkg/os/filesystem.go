package os

import (
	"io"
	"os"
)

func IsDirectoryEmpty(name string) bool {
	//gosec:disable G304 -- This is a false positive
	f, err := os.Open(name)
	if err != nil {
		return false
	}
	//nolint:errcheck
	defer f.Close()

	_, err = f.Readdirnames(1)
	return err == io.EOF
}
