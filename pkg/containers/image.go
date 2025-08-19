package containers

import (
	"context"
	"fmt"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types/ref"
)

func FindNewDigest(r ref.Ref) (string, error) {
	ctx := context.Background()
	rc := regclient.New()
	defer func() {
		_ = rc.Close(ctx, r)
	}()

	r.Digest = ""
	manifest, err := rc.ManifestGet(ctx, r)
	if err != nil {
		return "", fmt.Errorf("failed to get manifest: %w", err)
	}
	digest := manifest.GetDescriptor().Digest

	return digest.String(), nil
}
