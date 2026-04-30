package main

import (
	"fmt"
	"io"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/uniget-org/cli/pkg/containers"
	"gitlab.com/uniget-org/cli/pkg/logging"
	myos "gitlab.com/uniget-org/cli/pkg/os"
	"gitlab.com/uniget-org/cli/pkg/tui"

	"gitlab.com/uniget-org/cli/pkg/tool"
)

var toolVersion string
var rawInspect bool

func initInspectCmd() {
	inspectCmd.Flags().StringVar(&toolVersion, "version", "", "Inspect a specific version of the tool")
	inspectCmd.Flags().BoolVar(&rawInspect, "raw", false, "Show raw contents")

	rootCmd.AddCommand(inspectCmd)
}

var inspectCmd = &cobra.Command{
	Use:     "inspect",
	Aliases: []string{},
	Short:   "Inspect tool",
	Long:    header + "\nInspect tools",
	Example: "" +
		"  Use regctl/jq/xargs/tar to display raw contents:\n" +
		"    regctl manifest get ghcr.io/uniget-org/tools/TOOL:latest --platform linux/amd64 --format raw-body \\\n" +
		"    | jq --raw-output '.layers[0].digest' \\\n" +
		"    | xargs -I{} regctl blob get ghcr.io/uniget-org/tools/TOOL {} \\\n" +
		"    | tar -tvz",
	Args: cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		var inspectTool *tool.Tool

		inspectToolImageTag := "main"
		if len(toolVersion) == 0 {
			assertMetadataFileExists()
			assertMetadataIsLoaded()

			inspectTool, err = tools.GetByName(args[0])
			if err != nil {
				return fmt.Errorf("error getting tool %s", args[0])
			}
			inspectTool.ReplaceVariables(viper.GetString("prefix")+viper.GetString("target"), arch, altArch)

		} else {
			inspectTool = &tool.Tool{
				Name:    args[0],
				Version: toolVersion,
			}
			inspectToolImageTag = toolVersion
		}

		logging.Info.Printfln("Inspecting %s %s\n", inspectTool.Name, inspectTool.Version)
		registries, repositories := inspectTool.GetSourcesWithFallback(registry, imageRepository)
		toolRef, err := containers.FindToolRef(registries, repositories, inspectTool.Name, inspectToolImageTag)
		if err != nil {
			return fmt.Errorf("error finding tool %s:%s: %s", inspectTool.Name, inspectTool.Version, err)
		}
		effectivePathRewriteRules := pathRewriteRules
		if rawInspect {
			effectivePathRewriteRules = []tool.PathRewrite{}
		}

		var progressPrinter *pterm.ProgressbarPrinter
		progressReader := tui.NewProgressReader(nil, nil)
		if myos.IsTty() && !viper.GetBool("debug") && !viper.GetBool("trace") {
			progressPrinter, err = pterm.DefaultProgressbar.WithTitle("Downloading").WithTotal(0).WithRemoveWhenDone().Start()
			if err != nil {
				panic(err)
			}
			progressReader = tui.NewProgressReader(
				func(n int64) {
					progressPrinter.Total = int(n)
				},
				func(n int64) {
					progressPrinter.Add(int(n))
				},
			)
			//nolint:errcheck
			defer progressPrinter.Stop()
		}

		err = toolCache.Get(toolRef, progressReader, func(reader io.ReadCloser) error { return nil })
		if err != nil {
			return fmt.Errorf("unable to get image: %s", err)
		}
		var files []string
		err = toolCache.Get(toolRef, progressReader, func(reader io.ReadCloser) error {
			files, err = inspectTool.Inspect(cmd.OutOrStdout(), reader, effectivePathRewriteRules)
			if err != nil {
				return fmt.Errorf("unable to inspect %s: %s", inspectTool.Name, err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("unable to inspect image: %s", err)
		}

		for _, file := range files {
			//nolint:errcheck
			fmt.Fprintln(logging.OutputWriter, file)
		}

		return nil
	},
}
