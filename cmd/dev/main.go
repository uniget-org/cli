package main

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/pkg/logging"
)

var (
	version              = "dev"
	logLevel             = "info"
	debug                = false
	trace                = false
	unigetToolsDirectory = os.Getenv("HOME") + "/private/uniget/tools"
	unigetTools          *UnigetTools
	unigetToolsNames     []string
	rootCmd              = &cobra.Command{
		Use:          "uniget-dev",
		Version:      version,
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logging.OutputWriter = cmd.OutOrStdout()
			logging.ErrorWriter = cmd.ErrOrStderr()

			if trace {
				pterm.EnableDebugMessages()
				logging.Level = pterm.LogLevelTrace

			} else if debug {
				pterm.EnableDebugMessages()
				logging.Level = pterm.LogLevelDebug

			} else {
				pterm.DisableDebugMessages()
				logging.Level = pterm.LogLevelInfo
			}

			logging.Init()

			return nil
		},
	}
	platform         = "auto"
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
	pf := rootCmd.PersistentFlags()
	pf.StringVarP(&unigetToolsDirectory, "directory", "d", unigetToolsDirectory, "Directory to search for tools")
	pf.StringVarP(&platform, "platform", "p", platform, "Git platform (auto, github, gitlab)")
	pf.StringVar(&logLevel, "log-level", logLevel, "Log level (trace, debug, info, warning, error)")
	pf.BoolVar(&debug, "debug", debug, "Set log level to debug")
	pf.BoolVar(&trace, "trace", trace, "Set log level to trace")

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
