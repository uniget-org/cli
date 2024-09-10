package cache

import (
	"context"
	"fmt"
	"testing"

	"github.com/regclient/regclient/types/ref"
	"github.com/uniget-org/cli/pkg/containers"
)

var (
	registryAddress     = "127.0.0.1:5000"
	registryRepository  = "uniget-org/tools"
	registryImage       = "jq"
	registryTag         = "1.7.1"
	toolRef             = containers.NewToolRef(registryAddress, registryRepository, registryImage, registryTag)
	expectedLayerSha256 = "856b78ac4575c4cf8be9942e3b6100e27d150497cd90c2eb041aa7bc2017c300"
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

	rc := containers.GetRegclient()
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
	containers.StartRegistryWithCallback(registryAddress, func() {
		err := addTestData(registryAddress, "uniget-org/tools", "jq", "1.7.1")
		if err != nil {
			panic(err)
		}

		m.Run()
	})
}
