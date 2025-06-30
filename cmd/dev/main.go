package main

import (
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

	initEditCmd()
	initNewCmd()
}

func main() {
	pf := rootCmd.PersistentFlags()
	pf.StringVarP(&unigetToolsDirectory, "directory", "d", unigetToolsDirectory, "Directory to search for tools")

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
