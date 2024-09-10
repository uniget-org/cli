package containers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/distribution/distribution/v3/configuration"
	"github.com/distribution/distribution/v3/registry"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/inmemory"
)

func CreateRegistry(registryAddress string) (*registry.Registry, error) {
	// https://distribution.github.io/distribution/about/configuration/
	const distributionConfig = `
version: 0.1
log:
  accesslog:
    disabled: true
    formatter: text
  level: info
storage:
  inmemory:
`

	ctx := context.Background()

	config, err := configuration.Parse(bytes.NewReader([]byte(distributionConfig)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse distribution configuration: %s", err)
	}
	config.HTTP.Addr = registryAddress

	registry, err := registry.NewRegistry(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create registry: %s", err)
	}

	return registry, nil
}

func StartBackgroundRegistry(registryAddress string, doShutdown *bool) {
	registry, err := CreateRegistry(registryAddress)
	if err != nil {
		fmt.Printf("failed to create registry: %s", err)
		return
	}

	go func() {
		for {
			time.Sleep(1 * time.Second)
			if *doShutdown {
				err := registry.Shutdown(context.Background())
				if err != nil {
					fmt.Printf("failed to shutdown registry: %s", err)
				}
				return
			}
		}
	}()

	err = registry.ListenAndServe()
	if err != nil {
		fmt.Printf("failed to start registry: %s", err)
		return
	}
}

func StartRegistryWithCallback(registryAddress string, callback func()) {
	registry, err := CreateRegistry(registryAddress)
	if err != nil {
		fmt.Printf("failed to create registry: %s", err)
		return
	}

	go func() {
		startTimestamp := time.Now().Unix()
		for {
			url := fmt.Sprintf("http://%s/v2/", registryAddress)
			res, err := http.Get(url)
			if err == nil {
				if res.StatusCode == http.StatusOK {
					break
				}
			}

			if time.Now().Unix()-startTimestamp > 60 {
				fmt.Printf("timeout waiting for registry to start")
				return
			}

			time.Sleep(1 * time.Second)
		}

		callback()
		err := registry.Shutdown(context.Background())
		if err != nil {
			fmt.Printf("failed to shutdown registry: %s", err)
		}
	}()

	err = registry.ListenAndServe()
	if err != nil {
		fmt.Printf("failed to start registry: %s", err)
		return
	}
}
