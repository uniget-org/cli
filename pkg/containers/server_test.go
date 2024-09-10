package containers

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func testRegistryRequests(t *testing.T, registryAddress string) {
	url := fmt.Sprintf("http://%s/v2/", registryAddress)
	res, err := http.Get(url) // #nosec G107 -- This is only a test and registryAddress is hardcoded
	if err != nil {
		t.Errorf("failed to get response: %s", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, res.StatusCode)
	}
}

func TestStartBackgroundRegistry(t *testing.T) {
	var registryAddress = "127.0.0.1:5001"
	var doShutdown = false
	go StartBackgroundRegistry(registryAddress, &doShutdown)

	time.Sleep(5 * time.Second)
	testRegistryRequests(t, registryAddress)

	doShutdown = true
}

func TestStartRegistryWithCallback(t *testing.T) {
	var registryAddress = "127.0.0.1:5001"
	StartRegistryWithCallback(registryAddress, func() {
		testRegistryRequests(t, registryAddress)
	})
}
