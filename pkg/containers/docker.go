package containers

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/uniget-org/cli/pkg/archive"
)

func GetFirstLayerFromDockerImage(cli *client.Client, ref string) ([]byte, error) {
	shaString, err := GetFirstLayerShaFromRegistry(ref)
	if err != nil {
		panic(err)
	}
	sha := shaString[7:]

	image, err := ReadDockerImage(cli, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %s", err)
	}

	layerGzip, err := UnpackLayerFromDockerImage(image, sha)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack layer: %s", err)
	}

	layer, err := archive.Gunzip(layerGzip)
	if err != nil {
		return nil, fmt.Errorf("failed to gunzip layer: %s", err)
	}

	return layer, nil
}

func PullDockerImage(cli *client.Client, ref string) error {
	ctx := context.Background()

	events, err := cli.ImagePull(ctx, ref, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %s", err)
	}
	defer events.Close()
	_, err = io.Copy(io.Discard, events)
	if err != nil {
		return fmt.Errorf("failed to read events: %s", err)
	}

	return nil
}

func CheckDockerImageExists(cli *client.Client, ref string) bool {
	ctx := context.Background()

	_, _, err := cli.ImageInspectWithRaw(ctx, ref)
	return err == nil
}

func ReadDockerImage(cli *client.Client, ref string) ([]byte, error) {
	ctx := context.Background()
	
	err := PullDockerImage(cli, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to pull image: %s", err)
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