package cache

import (
	"context"
	"fmt"
	"io"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/config"
	rref "github.com/regclient/regclient/types/ref"
	"gitlab.com/uniget-org/cli/pkg/containers"
	"gitlab.com/uniget-org/cli/pkg/logging"
)

type NoneCache struct{}

func NewNoneCache() *NoneCache {
	return &NoneCache{}
}

func (c *NoneCache) Get(tool *containers.ToolRef, callback func(reader io.ReadCloser) error) error {
	ctx := context.Background()

	r, err := rref.New(tool.String())
	if err != nil {
		return fmt.Errorf("failed to create reference for %s: %w", tool, err)
	}

	rcOpts := []regclient.Opt{}
	rcOpts = append(rcOpts, regclient.WithUserAgent("uniget"))
	rcOpts = append(rcOpts, regclient.WithDockerCreds())
	rcOpts = append(rcOpts, regclient.WithConfigHost(config.Host{
		Name: "127.0.0.1:5000",
		TLS:  config.TLSDisabled,
	}))
	rc := regclient.New(rcOpts...)
	//nolint:errcheck
	defer rc.Close(ctx, r)

	logging.Debugf("NoneCache: Pulling %s", r)

	err = containers.GetFirstLayerFromRegistry(ctx, rc, r, func(reader io.ReadCloser) error {
		err := callback(reader)
		if err != nil {
			return fmt.Errorf("failed to execute callback: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to get layer for ref %s: %w", tool, err)
	}

	return nil
}
