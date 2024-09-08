package containers

import (
	"bufio"
	"bytes"
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/images/archive"
	"github.com/containerd/platforms"
	uarchive "github.com/uniget-org/cli/pkg/archive"
)

func GetContainerdClient() (*containerd.Client, error) {
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		return nil, fmt.Errorf("failed to create containerd client: %s", err)
	}
	return client, nil
}

func GetFirstLayerFromContainerdImage(client *containerd.Client, ref *ToolRef) ([]byte, error) {
	shaString, err := GetFirstLayerShaFromRegistry(ref)
	if err != nil {
		panic(err)
	}
	sha := shaString[7:]

	imageData, err := ReadContainerdImage(client, ref.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %s", err)
	}

	layerGzip, err := UnpackLayerFromDockerImage(imageData, sha)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack layer: %s", err)
	}

	layer, err := uarchive.Gunzip(layerGzip)
	if err != nil {
		return nil, fmt.Errorf("failed to gunzip layer: %s", err)
	}

	return layer, nil
}

func CheckContainerdImageExists(client *containerd.Client, ref string) bool {
	ctx := context.Background()

	_, err := client.GetImage(ctx, ref)
	return err == nil
}

func PullContainerdImage(client *containerd.Client, ref string) error {
	ctx := context.Background()

	_, err := client.Pull(ctx, ref)
	if err != nil {
		return fmt.Errorf("failed to pull image: %s", err)
	}

	return nil
}

func ReadContainerdImage(client *containerd.Client, ref string) ([]byte, error) {
	ctx := context.Background()

	err := PullContainerdImage(client, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to pull image: %s", err)
	}

	var imageBuffer bytes.Buffer
	writer := bufio.NewWriter(&imageBuffer)
	is := client.ImageService()
	err = client.Export(ctx, writer, archive.WithImage(is, ref), archive.WithPlatform(platforms.DefaultStrict()))
	if err != nil {
		return nil, fmt.Errorf("failed to export image: %s", err)
	}
	imageData := imageBuffer.Bytes()

	return imageData, nil
}
