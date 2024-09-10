package containers

import (
	"context"
	"fmt"
	"testing"

	"github.com/regclient/regclient/types/ref"
)

var (
	registryAddress       = "127.0.0.1:5000"
	registryRepository    = "uniget-org/tools"
	registryImage         = "jq"
	registryTag           = "1.7.1"
	toolRef               = NewToolRef(registryAddress, registryRepository, registryImage, registryTag)
	expectedLayerGzSha256 = "ba8f179e01e11eb5a78d8d8c2c85ce72beaece1ce32f366976d30e7f1b161eae"
)

func addTestData(registryAddress, registryRepository, registryImage, registryTag string) error {
	ctx := context.Background()
	rSrc, err := ref.New(fmt.Sprintf("%s/%s/%s:%s", "ghcr.io", registryRepository, registryImage, registryTag))
	if err != nil {
		return err
	}
	rTgt, err := ref.New(fmt.Sprintf("%s/%s/%s:%s", registryAddress, registryRepository, registryImage, registryTag))
	if err != nil {
		return err
	}

	rc := GetRegclient()
	defer rc.Close(ctx, rSrc)
	defer rc.Close(ctx, rTgt)

	err = rc.ImageCopy(ctx, rSrc, rTgt)
	if err != nil {
		return err
	}

	return nil
}

func TestMain(m *testing.M) {
	var registryAddress = "127.0.0.1:5000"
	StartRegistryWithCallback(registryAddress, func() {
		err := addTestData(registryAddress, "uniget-org/tools", "jq", "1.7.1")
		if err != nil {
			panic(err)
		}

		m.Run()
	})
}
