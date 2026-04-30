package cache

import (
	"io"

	"gitlab.com/uniget-org/cli/pkg/containers"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

type Cache interface {
	Get(tool *containers.ToolRef, p tui.ProgressReader, callback func(reader io.ReadCloser) error) error
}
