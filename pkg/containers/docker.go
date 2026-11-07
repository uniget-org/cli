package containers

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/client"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

func GetDockerClient() (*client.Client, error) {
	cli, err := client.New(client.FromEnv)
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

	ping, err := cli.Ping(context.Background(), client.PingOptions{})
	if err != nil {
		return false
	}

	return ping.APIVersion != ""
}

func GetFirstLayerFromDockerImage(cli *client.Client, ref *ToolRef, p tui.ProgressReader, callback func(reader io.ReadCloser) error) error {
	logging.Tracef("Getting first layer for %s using docker", ref)

	shaString, err := GetFirstLayerShaFromRegistry(ref)
	if err != nil {
		return fmt.Errorf("failed to get first layer sha: %s", err)
	}
	sha := shaString[7:]

	err = ReadDockerImage(cli, ref.String(), p, func(reader io.ReadCloser) error {
		err := UnpackLayerFromDockerImage(reader, sha, func(blob io.ReadCloser) error {
			layerReader, err := NewLayerReader(blob)
			if err != nil {
				return fmt.Errorf("failed to create layer reader: %s", err)
			}

			err = callback(layerReader)
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

	events, err := cli.ImagePull(ctx, ref, client.ImagePullOptions{})
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

func ReadDockerImage(cli *client.Client, ref string, p tui.ProgressReader, callback func(reader io.ReadCloser) error) error {
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

	p.SetTotal(imageInspect.Size)
	p.SetReader(reader)
	err = callback(p)
	if err != nil {
		return fmt.Errorf("failed to execute callback: %w", err)
	}

	return nil
}

// dockerSaveManifestEntry mirrors the entries of the manifest.json file
// contained in docker save archives.
type dockerSaveManifestEntry struct {
	Config string   `json:"Config"`
	Layers []string `json:"Layers"`
}

// UnpackLayerFromDockerImage extracts the layer blob with digest sha256 from
// a docker save or containerd export archive and hands its content to the
// callback.
//
// Older engines store the layer exactly as referenced by the registry manifest
// (typically gzip). Docker Engine >= v29 stores layers uncompressed and
// recomputes the digest, so if the digest cannot be found, the layer path for
// the archive is resolved from the archive's manifest.json instead.
func UnpackLayerFromDockerImage(buffer io.ReadCloser, sha256 string, callback func(reader io.ReadCloser) error) error {
	data, err := io.ReadAll(buffer)
	if err != nil {
		return fmt.Errorf("UnpackLayerFromDockerImage(): failed to read buffer: %s", err)
	}
	//nolint:errcheck
	defer buffer.Close()

	tarReader := tar.NewReader(bytes.NewReader(data))

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

	// The requested digest is not present in the archive. Resolve the layer
	// path from manifest.json as stored by Docker Engine >= v29.
	layerPath, ok := getLayerPathFromManifest(data)
	if !ok {
		return fmt.Errorf("failed to extract layer %s", sha256)
	}

	tarReader = tar.NewReader(bytes.NewReader(data))
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break

		} else if err != nil {
			return fmt.Errorf("UnpackLayerFromDockerImage(): failed to find next item in tar: %s", err)
		}

		if header.Name == layerPath {
			switch header.Typeflag {
			case tar.TypeReg:
				err = callback(io.NopCloser(io.LimitReader(tarReader, header.Size)))
				if err != nil {
					return fmt.Errorf("UnpackLayerFromDockerImage(): failed to execute callback: %w", err)
				}
				return nil
			}
		}
	}

	return fmt.Errorf("failed to extract layer %s", sha256)
}

// getLayerPathFromManifest reads the manifest.json file from a docker save
// archive and returns the path of the first layer entry.
func getLayerPathFromManifest(data []byte) (string, bool) {
	tarReader := tar.NewReader(bytes.NewReader(data))

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break

		} else if err != nil {
			return "", false
		}

		if header.Name == "manifest.json" && header.Typeflag == tar.TypeReg {
			content, err := io.ReadAll(io.LimitReader(tarReader, header.Size))
			if err != nil {
				return "", false
			}

			entries := []dockerSaveManifestEntry{}
			err = json.Unmarshal(content, &entries)
			if err != nil {
				return "", false
			}

			for _, entry := range entries {
				if len(entry.Layers) > 0 && entry.Layers[0] != "" {
					return entry.Layers[0], true
				}
			}

			return "", false
		}
	}

	return "", false
}

// NewLayerReader wraps the content of a layer blob. Blobs stored by older
// engines and containerd exports are gzip compressed, while blobs saved by
// Docker Engine >= v29 are uncompressed tar archives. In both cases the
// returned reader yields the uncompressed layer content.
func NewLayerReader(blob io.Reader) (io.ReadCloser, error) {
	first := make([]byte, 2)
	_, err := io.ReadFull(blob, first)
	if err != nil {
		return nil, fmt.Errorf("failed to read layer prefix: %w", err)
	}

	if first[0] == 0x1f && first[1] == 0x8b {
		gzipReader, err := gzip.NewReader(io.MultiReader(bytes.NewReader(first), blob))
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}

		return io.NopCloser(gzipReader), nil
	}

	return io.NopCloser(io.MultiReader(bytes.NewReader(first), blob)), nil
}

func ListDockerImagesByPrefix(cli *client.Client, prefix string) ([]image.Summary, error) {
	ctx := context.Background()
	images, err := cli.ImageList(ctx, client.ImageListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	var filtered []image.Summary
	for _, img := range images.Items {
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
	_, err := cli.ImageRemove(ctx, ref, client.ImageRemoveOptions{Force: true, PruneChildren: true})
	if err != nil {
		return fmt.Errorf("failed to remove image %s: %w", ref, err)
	}
	return nil
}
