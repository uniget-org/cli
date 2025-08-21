package cache

import (
	"io"

	"github.com/uniget-org/cli/pkg/containers"
)

type Cache interface {
	Get(tool *containers.ToolRef) (io.ReadCloser, error)
}
