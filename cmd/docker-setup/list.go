package main

import (
	"github.com/nicholasdille/docker-setup/pkg/tool"
	"github.com/spf13/cobra"
)

var installedOnly bool

func initListCmd() {
	listCmd.Flags().BoolVar(&installedOnly, "installed", false, "List only installed tools")

	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "get"},
	Short:   "List tools",
	Long:    header + "\nList tools",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		assertMetadataFileExists()
		assertMetadataIsLoaded()

		if installedOnly {
			var installedTools tool.Tools
			for index := range tools.Tools {
				tools.Tools[index].ReplaceVariables(prefix+"/"+target, arch, altArch)
				tools.Tools[index].GetMarkerFileStatus(prefix + "/" + cacheDirectory)
				tools.Tools[index].GetBinaryStatus()
				tools.Tools[index].GetVersionStatus()

				if tools.Tools[index].Status.VersionMatches {
					installedTools.Tools = append(installedTools.Tools, tools.Tools[index])
				}
			}
			installedTools.List()

		} else {
			tools.List()
		}

		return nil
	},
}
