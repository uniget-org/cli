package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/charmbracelet/x/term"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/internal/constants"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"go.yaml.in/yaml/v3"
)

var listInstalledOnly bool
var listUpgradableOnly bool
var listOutput string

func initListCmd() {
	listCmd.Flags().BoolVar(&listInstalledOnly, "installed", false, "List only installed tools")
	listCmd.Flags().BoolVar(&listUpgradableOnly, "upgradable", false, "List only upgradable tools")
	listCmd.Flags().StringVarP(&listOutput, "output", "o", "pretty", "Output options: pretty, json, yaml")

	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use: "list",
	Aliases: []string{
		"l",
		"get",
	},
	Short:   "List tools",
	Long:    constants.Header + "\nList tools",
	GroupID: "tool",
	Args:    cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		tools, err = configuration.EnsureMetadata()
		if err != nil {
			return fmt.Errorf("error ensuring metadata: %s", err)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		type listTool struct {
			Name             string `json:"name" yaml:"name"`
			InstalledVersion string `json:"installed_version" yaml:"installed_version"`
			AvailableVersion string `json:"available_version" yaml:"available_version"`
			Description      string `json:"description" yaml:"description"`
		}

		var listTools []listTool

		if listInstalledOnly {
			for index := range tools.Tools {
				err := tools.Tools[index].UpdateStatus(
					configuration.Prefix,
					configuration.Target,
					configuration.GetCacheDirectory(),
					configuration.Arch,
					configuration.AltArch,
				)
				if err != nil {
					return fmt.Errorf("failed to update status for tool %s: %s", tools.Tools[index].Name, err)
				}

				if tools.Tools[index].IsInstalled() {
					listTools = append(listTools, listTool{
						Name:             tools.Tools[index].Name,
						InstalledVersion: tools.Tools[index].Status.Version,
						AvailableVersion: tools.Tools[index].Version,
						Description:      tools.Tools[index].Description,
					})
				}
			}

		} else if listUpgradableOnly {
			for index := range tools.Tools {
				err := tools.Tools[index].UpdateStatus(
					configuration.Prefix,
					configuration.Target,
					configuration.GetCacheDirectory(),
					configuration.Arch,
					configuration.AltArch,
				)
				if err != nil {
					return fmt.Errorf("failed to update status for tool %s: %s", tools.Tools[index].Name, err)
				}

				if tools.Tools[index].IsUpgradable() {
					listTools = append(listTools, listTool{
						Name:             tools.Tools[index].Name,
						InstalledVersion: tools.Tools[index].Status.Version,
						AvailableVersion: tools.Tools[index].Version,
						Description:      tools.Tools[index].Description,
					})
				}
			}

		} else {
			for index := range tools.Tools {
				if tools.Tools[index].Lifecycle.RemovedWithReason != "" || tools.Tools[index].Lifecycle.RenamedTo != "" {
					continue
				}

				err := tools.Tools[index].UpdateStatus(
					configuration.Prefix,
					configuration.Target,
					configuration.GetCacheDirectory(),
					configuration.Arch,
					configuration.AltArch,
				)
				if err != nil {
					return fmt.Errorf("failed to update status for tool %s: %s", tools.Tools[index].Name, err)
				}

				listTools = append(listTools, listTool{
					Name:             tools.Tools[index].Name,
					InstalledVersion: tools.Tools[index].Status.Version,
					AvailableVersion: tools.Tools[index].Version,
					Description:      tools.Tools[index].Description,
				})
			}
		}

		switch listOutput {
		case "pretty":
			tableColumnConfig := []table.ColumnConfig{
				{
					Number:   1,
					WidthMax: 5,
				},
				{
					Number:   2,
					WidthMax: 30,
				},
				{
					Number:   3,
					WidthMax: 10,
				},
				{
					Number:   4,
					WidthMax: 20,
				},
				{
					Number:   5,
					WidthMax: 35,
				},
			}
			separatorWidth := (len(tableColumnConfig) - 1) * 2
			requiredWidth := separatorWidth
			for _, col := range tableColumnConfig {
				requiredWidth += col.WidthMax
			}
			logging.Debugf("Required width: %d", requiredWidth)

			terminalWidth, _, err := term.GetSize(os.Stdout.Fd())
			if err != nil {
				if columns, envErr := strconv.Atoi(os.Getenv("COLUMNS")); envErr == nil && columns > 0 {
					terminalWidth = columns
				} else {
					terminalWidth = requiredWidth + 1
				}
			}
			logging.Debugf("Terminal width: %d", terminalWidth)

			remainingDescriptionColumnWidth := terminalWidth - (requiredWidth - tableColumnConfig[4].WidthMax + 1)
			logging.Debugf("Remaining column width description: %d", remainingDescriptionColumnWidth)
			if remainingDescriptionColumnWidth < tableColumnConfig[4].WidthMax {
				return fmt.Errorf("terminal width is too small to display the table")
			}
			logging.Debugf("Setting description column width to: %d", remainingDescriptionColumnWidth)
			tableColumnConfig[4].WidthMax = remainingDescriptionColumnWidth

			listTable := table.NewWriter()
			listTable.SetOutputMirror(cmd.OutOrStdout())
			listTable.SetColumnConfigs(tableColumnConfig)
			listTable.SetRowPainter(func(row table.Row) text.Colors {
				defaultColor := text.Colors{
					text.FgWhite,
				}

				if row[2] == row[3] {
					defaultColor = text.Colors{
						text.FgGreen,
					}

				} else if row[2] != "" && row[2] != row[3] {
					defaultColor = text.Colors{
						text.FgRed,
					}
				}

				return defaultColor
			})
			listTable.Style().Options.DrawBorder = false
			listTable.Style().Options.SeparateColumns = false
			listTable.Style().Options.SeparateFooter = false
			listTable.Style().Options.SeparateHeader = false
			listTable.Style().Options.SeparateRows = false

			listTable.AppendHeader(table.Row{"#", "Name", "Installed", "Version", "Description"})

			for index, tool := range listTools {
				row := table.Row{index + 1, tool.Name, tool.InstalledVersion, tool.AvailableVersion, tool.Description}
				listTable.AppendRow(
					row,
					table.RowConfig{
						//
					},
				)
			}

			listTable.Render()
		case "json":
			data, err := json.Marshal(listTools)
			if err != nil {
				return fmt.Errorf("failed to marshal to json: %s", err)
			}
			fmt.Println(string(data))
		case "yaml":
			yamlEncoder := yaml.NewEncoder(cmd.OutOrStdout())
			yamlEncoder.SetIndent(2)
			defer func() {
				err := yamlEncoder.Close()
				if err != nil {
					logging.Warning.Printfln("failed to close yaml encoder: %s", err)
				}
			}()
			err := yamlEncoder.Encode(listTools)
			if err != nil {
				return fmt.Errorf("failed to encode yaml: %s", err)
			}
		default:
			return fmt.Errorf("invalid output format: %s", listOutput)
		}

		return nil
	},
}
