package main

import (
	"context"
	"fmt"
	"io"

	rref "github.com/regclient/regclient/types/ref"
	"gitlab.com/uniget-org/cli/pkg/containers"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

// fetchLayer pulls the first layer of the tool image and hands it to the
// callback. Previously, this was implemented as cache backends in pkg/cache.
func fetchLayer(tool *containers.ToolRef, p tui.ProgressReader, callback func(reader io.ReadCloser) error) error {
	ctx := context.Background()

	r, err := rref.New(tool.String())
	if err != nil {
		return fmt.Errorf("failed to create reference for %s: %w", tool, err)
	}

	rc := containers.GetRegclient()
	//nolint:errcheck
	defer rc.Close(ctx, r)

	logging.Debugf("Pulling %s", r)

	err = containers.GetFirstLayerFromRegistry(ctx, rc, r, p, func(reader io.ReadCloser) error {
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
