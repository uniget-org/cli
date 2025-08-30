package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/uniget-org/cli/pkg/containers"
	"github.com/uniget-org/cli/pkg/parse"
)

var (
	bumpDockerfile = "Dockerfile"
)

func initBumpCmd() {
	bumpDockerfileCmd.Flags().StringVarP(&bumpDockerfile, "file", "f", bumpDockerfile, "Path to Dockerfile")

	bumpCmd.AddCommand(bumpDockerfileCmd)
	rootCmd.AddCommand(bumpCmd)
}

var bumpCmd = &cobra.Command{
	Use:   "bump",
	Short: "Bump tool versions",
	Long:  header + "\nBump tool versions",
	Args:  cobra.NoArgs,
}

var bumpDockerfileCmd = &cobra.Command{
	Use:   "dockerfile",
	Short: "Bump image references in a Dockerfile",
	Long:  header + "\nBump image references in a Dockerfile",
	Args:  cobra.NoArgs,
	RunE:  processDockerfile,
}

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

func processDockerfile(cmd *cobra.Command, args []string) error {
	assertMetadataFileExists()
	assertMetadataIsLoaded()

	file, err := SlurpFile(bumpDockerfile)
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

	stat, err := os.Stat(bumpDockerfile)
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %w", bumpDockerfile, err)
	}
	err = os.WriteFile(bumpDockerfile, file, stat.Mode())
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", bumpDockerfile, err)
	}

	return nil
}
