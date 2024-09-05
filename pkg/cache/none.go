package cache

import (
	"context"

	"github.com/regclient/regclient"
	rref "github.com/regclient/regclient/types/ref"
	"github.com/uniget-org/cli/pkg/containers"
)

type NoneCache struct {}

func NewNoneCache() *NoneCache {
	return &NoneCache{}
}

func (c *NoneCache) Get(tool *ToolRef) ([]byte, error) {
	ctx := context.Background()

	r, err := rref.New(tool.String())
	if err != nil {
		panic(err)
	}

	rcOpts := []regclient.Opt{}
	rcOpts = append(rcOpts, regclient.WithUserAgent("uniget"))
	rcOpts = append(rcOpts, regclient.WithDockerCreds())
	rc := regclient.New(rcOpts...)
	defer rc.Close(ctx, r)

	layer, err := containers.GetFirstLayerFromRegistry(ctx, rc, r)
	if err != nil {
		panic(err)
	}

	return layer, nil
}