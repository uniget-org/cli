package archive

import (
	"io"

	"github.com/ulikunitz/xz"
)

func Untarxz(dst string, r io.Reader) error {
	xzr, err := xz.NewReader(r)
    if err != nil {
        return err
    }

	err = Untar(dst, xzr)
	if err != nil {
		return err
	}

	return nil
}