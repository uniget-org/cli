package cache

import (
	"io"

	"github.com/uniget-org/cli/pkg/containers"
)

type Cache interface {
	Get(tool *containers.ToolRef, callback func(reader io.ReadCloser) error) error
}
