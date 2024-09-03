package main

import (
	"fmt"
	"os"

	"github.com/containerd/containerd"
	uarchive "github.com/uniget-org/cli/pkg/archive"
	"github.com/uniget-org/cli/pkg/containers"
)

func GetFirstLayerFromContainerdImage(client *containerd.Client, ref string) ([]byte, error) {
	shaString, err := containers.GetFirstLayerShaFromRegistry(ref)
	if err != nil {
		panic(err)
	}
	sha := shaString[7:]
	
	imageData, err := containers.GetContainerdImage(client, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %s", err)
	}

	layerGzip, err := containers.UnpackLayerFromDockerImage(imageData, sha)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack layer: %s", err)
	}

	layer, err := uarchive.Gunzip(layerGzip)
	if err != nil {
		return nil, fmt.Errorf("failed to gunzip layer: %s", err)
	}

	return layer, nil
}
  
func main() {
	registry := "ghcr.io"
	repository := "uniget-org/tools"
	tool := "jq"
	tag := "latest"
	ref := fmt.Sprintf("%s/%s/%s:%s", registry, repository, tool, tag)

	namespace := "uniget"
	
	// TODO: Support equivalent of `ctr --address=...`
	address := "/run/containerd/containerd.sock"
	client, err := containerd.New(address, containerd.WithDefaultNamespace(namespace))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	layer, err := GetFirstLayerFromContainerdImage(client, ref)
	if err != nil {
		panic(err)
	}
	
	err = os.WriteFile(fmt.Sprintf("%s-%s.tar", tool, tag), layer, 0644)
	if err != nil {
		panic(err)
	}
}