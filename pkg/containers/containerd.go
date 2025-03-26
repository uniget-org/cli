package containers

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/images/archive"
	"github.com/containerd/platforms"
)

func GetContainerdClient() (*containerd.Client, error) {
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		return nil, fmt.Errorf("failed to create containerd client: %s", err)
	}
	return client, nil
}

func ContainerdIsAvailable() bool {
	client, err := GetContainerdClient()
	if err != nil {
		return false
	}
	//nolint:errcheck
	defer client.Close()

	version, err := client.Version(context.Background())
	if err != nil {
		return false
	}

	return version.Version != ""
}

func GetFirstLayerFromContainerdImage(client *containerd.Client, ref *ToolRef) ([]byte, error) {
	shaString, err := GetFirstLayerShaFromRegistry(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to get first layer sha: %s", err)
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

	reader, err := gzip.NewReader(bytes.NewReader(layerGzip))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %s", err)
	}
	//nolint:errcheck
	defer reader.Close()

	buffer, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read gzip: %s", err)
	}

	return buffer, nil
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
