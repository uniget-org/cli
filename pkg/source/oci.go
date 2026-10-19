package source

import (
	"context"
	"fmt"
	"io"
	"strings"

	rref "github.com/regclient/regclient/types/ref"
	"gitlab.com/uniget-org/cli/pkg/containers"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

type OciBackend struct {
}

func NewOciDownloader() *OciBackend {
	return &OciBackend{}
}

func IsOciRef(source *Source) bool {
	return strings.HasPrefix(source.Url, "oci://")
}

func (d *OciBackend) Get(source *Source, p tui.ProgressReader, callback func(reader io.ReadCloser) error) error {
	return d.GetFromRegistry(source, p, callback)
}

func (d *OciBackend) GetFromRegistry(source *Source, p tui.ProgressReader, callback func(reader io.ReadCloser) error) error {
	ctx := context.Background()

	ref := strings.TrimPrefix(source.Url, "oci://")
	r, err := rref.New(ref)
	if err != nil {
		return fmt.Errorf("failed to create reference for %s: %w", ref, err)
	}

	rc := containers.GetRegclient()
	//nolint:errcheck
	defer rc.Close(ctx, r)

	logging.Debugf("Pulling %s", r)
	err = containers.GetFirstLayerFromRegistryRaw(ctx, rc, r, p, func(reader io.ReadCloser) error {
		err := callback(reader)
		if err != nil {
			return fmt.Errorf("failed to execute callback: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to get layer for ref %s: %w", ref, err)
	}

	return nil
}
