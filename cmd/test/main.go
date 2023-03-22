package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nicholasdille/docker-setup/pkg/archive"
	"github.com/nicholasdille/docker-setup/pkg/containers"
	"github.com/nicholasdille/docker-setup/pkg/tool"

	_ "crypto/sha256"
	_ "crypto/sha512"
	"github.com/opencontainers/go-digest"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types"
	"github.com/regclient/regclient/types/ref"
	"github.com/regclient/regclient/types/manifest"
)

var version = "main"

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

	m, err := containers.GetPlatformManifest(ctx, rc, r, alt_arch)
	if err != nil {
		fmt.Printf("Failed to get manifest: %s\n", err)
		os.Exit(1)
	}
	err = ProcessLayers(ctx, rc, m, r, prefix)
	if err != nil {
		fmt.Printf("Failed to process layers: %s\n", err)
		os.Exit(1)
	}
}

func ProcessLayers(ctx context.Context, rc *regclient.RegClient, m manifest.Manifest, r ref.Ref, prefix string) error {
	if m.IsList() {
		return fmt.Errorf("Manifest is a list")
	}

	mi, ok := m.(manifest.Imager)
	if !ok {
		return fmt.Errorf("ERROR")
	}

	layers, err := mi.GetLayers()
	if err != nil {
		return fmt.Errorf("Failed to get layers: %s", err)
	}
	
	if len(layers) > 1 {
		return fmt.Errorf("Image must have exactly one layer but got %d", len(layers))
	}

	layer := layers[0]
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
		archive.ExtractTarGz(blob)

		return nil
	}
	
	// TODO: Test unknown media types
	return fmt.Errorf("Unknown media type encountered: %s", layer.MediaType)
}
