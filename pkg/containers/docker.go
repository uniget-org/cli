package containers

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/uniget-org/cli/pkg/logging"
)

func GetDockerClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %s", err)
	}
	return cli, nil
}

func DockerIsAvailable() bool {
	cli, err := GetDockerClient()
	if err != nil {
		return false
	}
	//nolint:errcheck
	defer cli.Close()

	ping, err := cli.Ping(context.Background())
	if err != nil {
		return false
	}

	return ping.APIVersion != ""
}

func GetFirstLayerFromDockerImage(cli *client.Client, ref *ToolRef, callback func(reader io.ReadCloser) error) error {
	logging.Tracef("Getting first layer for %s using docker", ref)

	shaString, err := GetFirstLayerShaFromRegistry(ref)
	if err != nil {
		return fmt.Errorf("failed to get first layer sha: %s", err)
	}
	sha := shaString[7:]

	err = ReadDockerImage(cli, ref.String(), func(reader io.ReadCloser) error {
		err := UnpackLayerFromDockerImage(reader, sha, func(reader io.ReadCloser) error {
			err = callback(reader)
			if err != nil {
				return fmt.Errorf("failed to execute callback: %w", err)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to unpack layer: %s", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to get image: %s", err)
	}

	return nil
}

func PullDockerImage(cli *client.Client, ref string) error {
	ctx := context.Background()

	events, err := cli.ImagePull(ctx, ref, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %s", err)
	}
	//nolint:errcheck
	defer events.Close()
	_, err = io.Copy(io.Discard, events)
	if err != nil {
		return fmt.Errorf("failed to read events: %s", err)
	}

	return nil
}

func CheckDockerImageExists(cli *client.Client, ref string) bool {
	ctx := context.Background()

	_, err := cli.ImageInspect(ctx, ref)
	return err == nil
}

func ReadDockerImage(cli *client.Client, ref string, callback func(reader io.ReadCloser) error) error {
	ctx := context.Background()

	err := PullDockerImage(cli, ref)
	if err != nil {
		return fmt.Errorf("failed to pull image: %s", err)
	}

	imageInspect, err := cli.ImageInspect(ctx, ref)
	if err != nil {
		return fmt.Errorf("failed to inspect image: %s", err)
	}
	imageID := imageInspect.ID

	reader, err := cli.ImageSave(ctx, []string{imageID})
	if err != nil {
		return fmt.Errorf("failed to save image: %s", err)
	}

	err = callback(reader)
	if err != nil {
		return fmt.Errorf("failed to execute callback: %w", err)
	}

	return nil
}

func UnpackLayerFromDockerImage(buffer io.ReadCloser, sha256 string, callback func(reader io.ReadCloser) error) error {
	tarReader := tar.NewReader(buffer)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break

		} else if err != nil {
			return fmt.Errorf("UnpackLayerFromDockerImage(): failed to find next item in tar: %s", err)
		}

		if header.Name == fmt.Sprintf("blobs/sha256/%s", sha256) {
			switch header.Typeflag {
			case tar.TypeReg:
				layerReader, err := gzip.NewReader(
					io.LimitReader(
						tarReader,
						header.Size,
					),
				)
				if err != nil {
					return fmt.Errorf("failed to create gzip reader: %w", err)
				}
				err = callback(io.NopCloser(layerReader))
				if err != nil {
					return fmt.Errorf("UnpackLayerFromDockerImage(): failed to execute callback: %w", err)
				}
				return nil
			}
		}

	}

	return fmt.Errorf("failed to extract layer %s", sha256)
}

func ListDockerImagesByPrefix(cli *client.Client, prefix string) ([]image.Summary, error) {
	ctx := context.Background()
	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	var filtered []image.Summary
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if strings.HasPrefix(tag, prefix) {
				filtered = append(filtered, img)
				break
			}
		}
	}
	return filtered, nil
}

func RemoveDockerImage(cli *client.Client, ref string) error {
	ctx := context.Background()
	_, err := cli.ImageRemove(ctx, ref, image.RemoveOptions{Force: true, PruneChildren: true})
	if err != nil {
		return fmt.Errorf("failed to remove image %s: %w", ref, err)
	}
	return nil
}
