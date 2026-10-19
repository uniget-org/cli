package source

import (
	"fmt"
	"io"

	"gitlab.com/uniget-org/cli/pkg/tui"
)

type Source struct {
	Url string
}

func NewBackendFromScheme(url *Source) (Downloader, error) {
	if IsFileRef(url) {
		return NewFileDownloader(), nil

	} else if IsWebRef(url) {
		return NewWebDownloader(), nil

	} else if IsOciRef(url) {
		return NewOciDownloader(), nil
	}

	return nil, fmt.Errorf("unsupported scheme for url: %s", url.Url)
}

type Downloader interface {
	Get(source *Source, p tui.ProgressReader, callback func(reader io.ReadCloser) error) error
}
