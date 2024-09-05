package main

import (
	"fmt"
	"os"

	"github.com/containerd/containerd"
	"github.com/uniget-org/cli/pkg/containers"
)
  
func main() {
	registry := "ghcr.io"
	repository := "uniget-org/tools"
	tool := "jq"
	tag := "latest"
	ref := fmt.Sprintf("%s/%s/%s:%s", registry, repository, tool, tag)

	namespace := "uniget"
	
	client, err := containerd.New("/run/containerd/containerd.sock", containerd.WithDefaultNamespace(namespace))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	layer, err := containers.GetFirstLayerFromContainerdImage(client, ref)
	if err != nil {
		panic(err)
	}
	
	err = os.WriteFile(fmt.Sprintf("%s-%s.tar", tool, tag), layer, 0644) // #nosec G306 -- just for testing
	if err != nil {
		panic(err)
	}
}