package archive

import (
	"compress/gzip"
	"io"
)

func Untargz(dst string, r io.Reader) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	err = Untar(dst, gzr)
	if err != nil {
		return err
	}

	return nil
}