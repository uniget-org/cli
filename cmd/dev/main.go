package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version              = "dev"
	unigetToolsDirectory = os.Getenv("HOME") + "/private/uniget/tools"
	unigetTools          *UnigetTools
	unigetToolsNames     []string
	rootCmd              = &cobra.Command{
		Use:          "uniget-dev",
		Version:      version,
		SilenceUsage: true,
	}
	gitForge         = "github"
	registryHost     = "ghcr.io"
	repositoryOwner  = "uniget-org"
	repositoryName   = "tools"
	repositoryPrefix = fmt.Sprintf("%s/%s", repositoryOwner, repositoryName)
	metadataTag      = "main"
	dockerTag        = metadataTag
)

func init() {
	unigetTools = NewUnigetTools(
		unigetToolsDirectory,
	)
	unigetTools.FindTools()
	unigetToolsNames = make([]string, 0, len(unigetTools.Tools))
	for k := range unigetTools.Tools {
		unigetToolsNames = append(unigetToolsNames, k)
	}

	initDebugCmd()
	initEditCmd()
	initManifestCmd()
	initMetadataCmd()
	initNewCmd()
}

func main() {
	pf := rootCmd.Flags()
	pf.StringVarP(&unigetToolsDirectory, "directory", "d", unigetToolsDirectory, "Directory to search for tools")
	pf.StringVarP(&gitForge, "forge", "f", gitForge, "Git forge (github, gitlab)")

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
