package containers

import (
	"bufio"
	"bytes"
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/images/archive"
	"github.com/containerd/platforms"
)

func GetContainerdImage(client *containerd.Client, ref string) ([]byte, error) {
	ctx := context.Background()

	image, err := client.Pull(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to pull image: %s", err)
	}

	var imageBuffer bytes.Buffer
	writer := bufio.NewWriter(&imageBuffer)
	is := client.ImageService()
	err = client.Export(ctx, writer, archive.WithImage(is, image.Name()), archive.WithPlatform(platforms.DefaultStrict()))
	if err != nil {
		return nil, fmt.Errorf("failed to export image: %s", err)
	}
	imageData := imageBuffer.Bytes()
	
	return imageData, nil
}