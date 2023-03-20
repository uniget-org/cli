package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nicholasdille/docker-setup/pkg/tool"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types/ref"
	"github.com/regclient/regclient/types/manifest"
)

func main() {
	arch := "x86_64"
	alt_arch := "amd64"
	target := "/usr/local"

	tools, err := tool.LoadFromFile("metadata.json")
	if err != nil {
		os.Exit(1)
	}

	tool, err := tools.GetByName("docker")
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

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(10 * time.Second))
	defer cancel()

	r, err := ref.New(image)
	if err != nil {
		return
	}

	rcOpts := []regclient.Opt{}
	rcOpts = append(rcOpts, regclient.WithUserAgent("docker-setup"))
	rcOpts = append(rcOpts, regclient.WithDockerCreds())
	rc := regclient.New(rcOpts...)
	defer rc.Close(ctx, r)

	m, err := rc.ManifestGet(ctx, r)
	if err != nil {
        fmt.Println(err)
		return
	}

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

					// TODO: Get manifest

					break
				}
			}
		}

	} else {
		fmt.Println("no list")

		if mi, ok := m.(manifest.Imager); ok {
			fmt.Printf("Type of mi: %T", mi)
			layers, err := mi.GetLayers()
			if err != nil {
				fmt.Println("Error getting layers")
				os.Exit(1)
			}
			fmt.Printf("Layer count: %d\n", len(layers))
		}
	}
}
