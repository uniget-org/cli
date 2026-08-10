package metadata

import (
	"fmt"
	"io"

	"github.com/google/safearchive/tar"
	"gitlab.com/uniget-org/cli/pkg/archive"
)

type Unpacker interface {
	Unpack(reader io.ReadCloser) error
}

type TarGzUnpacker struct{}

func NewTarGzUnpacker() Unpacker {
	return &TarGzUnpacker{}
}

func (u TarGzUnpacker) Unpack(upstreamReader io.ReadCloser) error {
	err := archive.Gunzip(upstreamReader, func(gunzipReader io.Reader) error {
		err := archive.Untar(io.NopCloser(gunzipReader), func(tarReader *tar.Reader, header *tar.Header) error {
			err := archive.CallbackExtractTarItem(tarReader, header)
			if err != nil {
				return fmt.Errorf("error extracting tar item %s: %s", header.Name, err)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("error untarring: %s", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error gunzipping: %s", err)
	}

	return nil
}

type NullUnpacker struct{}

func NewNullUnpacker() Unpacker {
	return &NullUnpacker{}
}

func (u NullUnpacker) Unpack(upstreamReader io.ReadCloser) error {
	return nil
}
