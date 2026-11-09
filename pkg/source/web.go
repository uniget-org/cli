package source

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"gitlab.com/uniget-org/cli/pkg/tui"
)

type WebBackend struct {
}

func NewWebDownloader() *WebBackend {
	return &WebBackend{}
}

func IsWebRef(source *Source) bool {
	return strings.HasPrefix(source.URL, "http://") || strings.HasPrefix(source.URL, "https://")
}

func (d *WebBackend) Get(source *Source, p tui.ProgressReader, callback func(reader io.ReadCloser) error) (err error) {
	resp, err := http.Get(source.URL) // #nosec G107 -- URL is configurable by the user
	if err != nil {
		return fmt.Errorf("failed to get %s: %w", source.URL, err)
	}
	//nolint:errcheck
	defer resp.Body.Close()

	return callback(resp.Body)
}
