package source

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gitlab.com/uniget-org/cli/pkg/tui"
)

type FileBackend struct {
}

func NewFileDownloader() *FileBackend {
	return &FileBackend{}
}

func IsFileRef(source *Source) bool {
	return strings.HasPrefix(source.Url, "file://")
}

func (d *FileBackend) Get(source *Source, p tui.ProgressReader, callback func(reader io.ReadCloser) error) (err error) {
	f, err := os.Open(strings.TrimPrefix(source.Url, "file://"))
	if err != nil {
		return fmt.Errorf("failed to open file %s: %s", source.Url, err)
	}
	//nolint:errcheck
	defer f.Close()

	return callback(f)
}
