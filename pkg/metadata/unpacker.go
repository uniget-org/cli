package metadata

import (
	"io"

	"github.com/google/safearchive/tar"
	"gitlab.com/uniget-org/cli/pkg/archive"
)

type Unpacker interface {
	Unpack(reader io.ReadCloser) error
}

type TarGzUnpacker struct {
}

func NewTarGzUnpacker() *Unpacker {
	var unpacker Unpacker = TarGzUnpacker{}

	return &unpacker
}

func (u TarGzUnpacker) Unpack(upstreamReader io.ReadCloser) error {
	err := archive.Gunzip(upstreamReader, func(gunzipReader io.Reader) error {
		err := archive.Untar(io.NopCloser(gunzipReader), func(tarReader *tar.Reader, header *tar.Header) error {
			err := archive.CallbackExtractTarItem(tarReader, header)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
