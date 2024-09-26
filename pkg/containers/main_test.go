package containers

import (
	"context"
	"testing"

	"github.com/regclient/regclient/types/ref"
)

var (
	registryAddress       = "127.0.0.1:5000"
	registryRepository    = "uniget-org/tools"
	registryImage         = "jq"
	registryTag           = "1.7.1"
	registryTags          = []string{"1.7.1", "latest"}
	toolRef               = NewToolRef(registryAddress, registryRepository, registryImage, registryTag)
	expectedLayerGzSha256 = "8f6a01445b829e82a8e6d5a66d646c1d884d0917df2c9a415a194cf273ac189d"
)

func copyImage(src, tgt ref.Ref) error {
	ctx := context.Background()
	rc := GetRegclient()
	defer rc.Close(ctx, src)
	defer rc.Close(ctx, tgt)

	err := rc.ImageCopy(ctx, src, tgt)
	if err != nil {
		return err
	}
	return nil
}

func addTestData() error {
	for _, tag := range registryTags {
		src := NewToolRef("ghcr.io", registryRepository, registryImage, tag)
		tgt := NewToolRef(registryAddress, registryRepository, registryImage, tag)
		err := copyImage(src.GetRef(), tgt.GetRef())
		if err != nil {
			return err
		}
	}

	return nil
}

func TestMain(m *testing.M) {
	var registryAddress = "127.0.0.1:5000"
	StartRegistryWithCallback(registryAddress, func() {
		err := addTestData()
		if err != nil {
			panic(err)
		}

		m.Run()
	})
}