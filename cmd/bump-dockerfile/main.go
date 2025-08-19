package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/uniget-org/cli/pkg/containers"
	"github.com/uniget-org/cli/pkg/parse"
	"github.com/uniget-org/cli/pkg/tool"
	"github.com/urfave/cli/v3"
)

var (
	registry    = []string{"ghcr.io"}
	repository  = []string{"uniget-org/tools"}
	metadataTag = "main"
	version     = "main"
)

func SlurpFile(filePath string) ([]byte, error) {
	f, err := os.Open(filePath) // #nosec G304 -- Data input
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %s", err)
	}
	defer func() {
		_ = f.Close()
	}()

	return io.ReadAll(f)
}

func main() {
	cmd := &cli.Command{
		Name:    "bump-dockerfile",
		Version: version,
		Usage:   "Update image references in a Dockerfile",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "input",
				Aliases: []string{"i"},
				Usage:   "Path to the input Dockerfile",
				Value:   "Dockerfile",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Path to the output Dockerfile",
				Value:   "Dockerfile",
			},
		},
		Action: process,
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func process(ctx context.Context, cmd *cli.Command) error {
	tools, err := tool.LoadMetadata(registry, repository, metadataTag)
	if err != nil {
		panic(err)
	}

	file, err := SlurpFile(cmd.String("input"))
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	reader := bytes.NewReader(file)
	imageRefs, err := parse.ExtractImageReferences(reader)
	if err != nil {
		return fmt.Errorf("failed to extract image references: %w", err)
	}

	for _, ref := range imageRefs.Refs {
		if ref.Registry == "ghcr.io" && ref.Repository[0:17] == "uniget-org/tools/" {
			toolName := ref.Repository[17:]
			tool, err := tools.GetByName(toolName)
			if err != nil {
				return fmt.Errorf("tool %s not found in metadata: %s", toolName, err)
			}

			refPattern := ref.Reference

			ref.Tag = tool.Version
			ref.Digest = ""
			ref.Reference = fmt.Sprintf("%s/%s:%s", ref.Registry, ref.Repository, ref.Tag)
			ref.Digest, err = containers.FindNewDigest(ref)
			if err != nil {
				return fmt.Errorf("failed to find new digest for %s: %w", ref, err)
			}

			refReplacement := fmt.Sprintf("%s/%s:%s@%s", ref.Registry, ref.Repository, ref.Tag, ref.Digest)

			if refPattern != refReplacement {
				fmt.Printf("Updating %s to %s with digest %s\n", tool.Name, tool.Version, ref.Digest)
				file = bytes.ReplaceAll(file, []byte(refPattern), []byte(refReplacement))
			}
		}
	}

	stat, err := os.Stat(cmd.String("output"))
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %w", cmd.String("output"), err)
	}
	err = os.WriteFile(cmd.String("output"), file, stat.Mode())
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", cmd.String("output"), err)
	}

	return nil
}
