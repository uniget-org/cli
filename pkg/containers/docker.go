package containers

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)


func GetDockerImage(cli *client.Client, ref string) ([]byte, error) {
	ctx := context.Background()
	
	events, err := cli.ImagePull(ctx, ref, image.PullOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to pull image: %s", err)
	}
	defer events.Close()
	_, err = io.Copy(io.Discard, events)
	if err != nil {
		return nil, fmt.Errorf("failed to read events: %s", err)
	}

	imageInspect, _, err := cli.ImageInspectWithRaw(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect image: %s", err)
	}
	imageID := imageInspect.ID

	reader, err := cli.ImageSave(ctx, []string{imageID})
	if err != nil {
		return nil, fmt.Errorf("failed to save image: %s", err)
	}
	defer reader.Close()
	
	buffer, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %s", err)
	}
	
	return buffer, nil
}

func UnpackLayerFromDockerImage(buffer []byte, sha256 string) ([]byte, error) {
	tarReader := tar.NewReader(bytes.NewReader(buffer))

	var layerBuffer []byte
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break

		} else if err != nil {
			return nil, fmt.Errorf("failed to find next item in tar: %s", err)
		}

		if header.Name != fmt.Sprintf("blobs/sha256/%s", sha256) {
			continue
		}

		switch header.Typeflag {
			case tar.TypeReg:
				layerBuffer, err = io.ReadAll(tarReader)
				if err != nil {
					return nil, fmt.Errorf("failed to read file from tar: %s", err)
				}
				return layerBuffer, nil
		}
	}

	return layerBuffer, fmt.Errorf("failed to extract layer %s", sha256)
}