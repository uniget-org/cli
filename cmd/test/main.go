package main

import (
	"archive/tar"
    "compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/nicholasdille/docker-setup/pkg/tool"

	_ "crypto/sha256"
	_ "crypto/sha512"
	"github.com/opencontainers/go-digest"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types"
	"github.com/regclient/regclient/types/ref"
	"github.com/regclient/regclient/types/manifest"
)

var arch = "x86_64"
var alt_arch = "amd64"
var prefix = "/"
var target = "usr/local"

func main() {
	if len(os.Args) == 1 {
		fmt.Printf("Expecting tool name as mandatory parameter\n")
		os.Exit(1)
	}
	toolName := os.Args[1]

	tools, err := tool.LoadFromFile("metadata.json")
	if err != nil {
		fmt.Printf("Failed to load metadata: %s\n", err)
		os.Exit(1)
	}

	tool, err := tools.GetByName(toolName)
	if err != nil {
		os.Exit(1)
	}
	tool.ReplaceVariables(target, arch, alt_arch)

	err = tool.GetBinaryStatus()
	if err != nil {
		fmt.Printf("Failed to get binary status: %s", err)
		os.Exit(1)
	}

	err = tool.GetVersionStatus()
	if err != nil {
		fmt.Printf("Failed to get version status: %s", err)
		os.Exit(1)
	}

	tool.Print()

	if tool.Status.VersionMatches {
		fmt.Printf("Nothing to do\n")
		os.Exit(0)
	}

	image := fmt.Sprintf("ghcr.io/nicholasdille/docker-setup/%s:main", tool.Name)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(30 * time.Second))
	defer cancel()

	r, err := ref.New(image)
	if err != nil {
		fmt.Printf("Failed to parse image name <%s>: %s\n", image, err)
		os.Exit(1)
	}

	rcOpts := []regclient.Opt{}
	rcOpts = append(rcOpts, regclient.WithUserAgent("docker-setup"))
	rcOpts = append(rcOpts, regclient.WithDockerCreds())
	rc := regclient.New(rcOpts...)
	defer rc.Close(ctx, r)

	m, err := rc.ManifestGet(ctx, r)
	if err != nil {
        fmt.Printf("Failed to get manifest: %s\n", err)
		os.Exit(1)
	}

	// TODO: Test manifest list with Docker media types
	// TODO: Test manifest list with OCI media types
	if m.IsList() {
		fmt.Println("list")

		if mi, ok := m.(manifest.Indexer); ok {
			manifests, err := mi.GetManifestList()
			if err != nil {
				fmt.Println("Error getting manifests")
				os.Exit(1)
			}
			fmt.Printf("Manifest count: %d\n", len(manifests))

			for i, manifest := range manifests {
				fmt.Printf("Manifest %d Platform %s\n", i, manifest.Platform.Architecture)
				if manifest.Platform.Architecture == alt_arch {
					fmt.Printf("Digest of manifest: %s\n", manifest.Digest)

					platformImage := fmt.Sprintf("%s@%s", image, manifest.Digest)
					r2, err := ref.New(platformImage)
					if err != nil {
						fmt.Printf("Failed to parse image name <%s>: %s\n", platformImage, err)
						os.Exit(1)
					}
					m2, err := rc.ManifestGet(ctx, r2)
					if err != nil {
						fmt.Printf("Failed to get manifest: %s\n", err)
						os.Exit(1)
					}
					if m2.IsList() {
						fmt.Printf("Manifest cannot be list again")
						os.Exit(1)
					}

					err = ProcessLayers(ctx, m2, r2, rc)
					if err != nil {
						fmt.Printf("Failed to process layers: %s\n", err)
						os.Exit(1)
					}

					break
				}
			}
		}

	// TODO: Test image with Docker media types
	// TODO: Test image with OCI media types
	} else {
		fmt.Println("no list")

		ProcessLayers(ctx, m, r, rc)
		if err != nil {
			fmt.Printf("Failed to process layers: %s\n", err)
			os.Exit(1)
		}
	}
}

func ProcessLayers(ctx context.Context, m manifest.Manifest, r ref.Ref, rc *regclient.RegClient) error {
	fmt.Printf("ProcessLayers()\n")

	if mi, ok := m.(manifest.Imager); ok {
		fmt.Printf("Type of mi: %T\n", mi)

		layers, err := mi.GetLayers()
		if err != nil {
			return fmt.Errorf("Failed to get layers: %s", err)
		}
		
		if len(layers) > 1 {
			return fmt.Errorf("Image must have exactly one layer but got %d", len(layers))
		}

		layer := layers[0]
		fmt.Printf("mediaType: %s\n", layer.MediaType)
		// TODO: Test known but unsupported media types
		if layer.MediaType == types.MediaTypeOCI1Layer || layer.MediaType == types.MediaTypeOCI1LayerZstd {
			return fmt.Errorf("Only layers with gzip compression are supported (not %s)", layer.MediaType)
		}
		if layer.MediaType == types.MediaTypeOCI1LayerGzip || layer.MediaType == types.MediaTypeDocker2LayerGzip  {
			
			d, err := digest.Parse(string(layer.Digest))
			if err != nil {
				return fmt.Errorf("Failed to parse digest %s: %s", layer.Digest, err)
			}

			blob, err := rc.BlobGet(ctx, r, types.Descriptor{Digest: d})
			if err != nil {
				return fmt.Errorf("Failed to get blob for digest %s: %s", layer.Digest, err)
			}
			defer blob.Close()

			//fmt.Printf("len of blob: %d\n", len(blob))
			//fmt.Printf("type of blob: %T\n", blob)

			os.Chdir(prefix)
			ExtractTarGz(blob)

			return nil
		}
		
		// TODO: Test unknown media types
		return fmt.Errorf("Unknown media type encountered: %s", layer.MediaType)	
	}

	return nil
}

// TODO: Check if https://github.com/mholt/archiver makes more sense
func ExtractTarGz(gzipStream io.Reader) {
    uncompressedStream, err := gzip.NewReader(gzipStream)
    if err != nil {
        fmt.Printf("ExtractTarGz: NewReader failed")
		os.Exit(1)
    }

    tarReader := tar.NewReader(uncompressedStream)

    for true {
        header, err := tarReader.Next()

        if err == io.EOF {
            break
        }

        if err != nil {
            fmt.Printf("ExtractTarGz: Next() failed: %s", err.Error())
			os.Exit(1)
        }

        switch header.Typeflag {
        case tar.TypeDir:
			if stat, err := os.Stat(header.Name); err == nil && ! stat.IsDir() {
				if err := os.Mkdir(header.Name, 0755); err != nil {
					fmt.Printf("ExtractTarGz: Mkdir() failed: %s", err.Error())
					os.Exit(1)
				}
			}

        case tar.TypeReg:
            outFile, err := os.Create(header.Name)
            if err != nil {
                fmt.Printf("ExtractTarGz: Create() failed: %s", err.Error())
				os.Exit(1)
            }
            if _, err := io.Copy(outFile, tarReader); err != nil {
                fmt.Printf("ExtractTarGz: Copy() failed: %s", err.Error())
				os.Exit(1)
            }
			outFile.Chmod(os.FileMode(header.Mode))
            outFile.Close()

        default:
            fmt.Printf("ExtractTarGz: uknown type: %s in %s", header.Typeflag, header.Name)
			os.Exit(1)
        }

    }
}
