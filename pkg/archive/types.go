package archive

import (
	"compress/gzip"
	"fmt"
	"io"

	"github.com/google/safearchive/tar"
)

func Untar(upstreamReader io.ReadCloser, callback func(reader io.Reader, header *tar.Header) error) error {
	reader := tar.NewReader(upstreamReader)
	//nolint:errcheck
	defer upstreamReader.Close()

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("failed to read next item: %s", err.Error())
		}

		err = callback(reader, header)
		if err != nil {
			return fmt.Errorf("failed to process item through callback: %s", err.Error())
		}
	}

	return nil
}

func Gunzip(upstreamReader io.ReadCloser, callback func(reader io.Reader) error) error {
	reader, err := gzip.NewReader(upstreamReader)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %s", err)
	}

	err = callback(reader)
	if err != nil {
		return fmt.Errorf("failed to execute callback: %s", err)
	}

	return nil
}
