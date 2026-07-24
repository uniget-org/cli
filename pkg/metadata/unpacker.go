package metadata

import "io"

type Unpacker interface {
	Unpack(reader io.ReadCloser) error
}

type TarGzUnpacker struct {
}

func NewTarGzUnpacker() *Unpacker {
	var unpacker Unpacker = TarGzUnpacker{}

	return &unpacker
}

func (u TarGzUnpacker) Unpack(reader io.ReadCloser) error {
	// Implement the logic to unpack a tar.gz file using the provided downloader and directory.
	// This is a placeholder implementation. You would need to fill in the actual unpacking logic.
	return nil
}
