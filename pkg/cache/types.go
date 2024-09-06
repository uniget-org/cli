package cache

import "github.com/uniget-org/cli/pkg/containers"

type Cache interface {
	Get(tool *containers.ToolRef) ([]byte, error)
}
