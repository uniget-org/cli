package main

import (
	"io"
	"os"

	"gitlab.com/uniget-org/cli/pkg/source"
	"gitlab.com/uniget-org/cli/pkg/source/cache"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

var pathPrefix = "./"
var ociRef = "ghcr.io/uniget-org/tools/uniget:latest"
var webUri = "https://gitlab.com/uniget-org/cli/-/releases/permalink/latest/downloads/uniget_Linux_x86_64.tar.gz"

func main() {
	oci()
	file()
	web()
}

func writeToFile(reader io.ReadCloser, filename string) error {
	f, err := os.Create(filename) // #nosec G304 - Test code, writing to a file is expected
	if err != nil {
		return err
	}
	//nolint:errcheck
	defer f.Close()
	_, err = io.Copy(f, reader)
	if err != nil {
		return err
	}

	return nil
}

func oci() {
	ociSource := &source.Source{
		Url: "oci://" + ociRef,
	}

	backend, err := source.NewOciDownloader(cache.CacheNone, nil)
	if err != nil {
		panic(err)
	}

	err = backend.Get(ociSource, tui.NewProgressReader(nil, nil), func(reader io.ReadCloser) error {
		//nolint:errcheck
		defer reader.Close()

		err = writeToFile(reader, pathPrefix+"oci")
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		panic(err)
	}
}

func file() {
	fileSource := &source.Source{
		Url: "file://" + pathPrefix + "oci",
	}

	src, err := source.NewFileDownloader(cache.CacheNone, nil)
	if err != nil {
		panic(err)
	}

	err = src.Get(fileSource, tui.NewProgressReader(nil, nil), func(reader io.ReadCloser) error {
		//nolint:errcheck
		defer reader.Close()

		err = writeToFile(reader, pathPrefix+"file")
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		panic(err)
	}
}

func web() {
	webSource := &source.Source{
		Url: webUri,
	}

	backend, err := source.NewWebDownloader(cache.CacheNone, nil)
	if err != nil {
		panic(err)
	}

	err = backend.Get(webSource, tui.NewProgressReader(nil, nil), func(reader io.ReadCloser) error {
		//nolint:errcheck
		defer reader.Close()

		err = writeToFile(reader, pathPrefix+"web")
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		panic(err)
	}
}
