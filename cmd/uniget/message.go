package main

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/uniget-org/cli/pkg/logging"
)

var find bool
var list bool

func initMessageCmd() {
	messageCmd.Flags().BoolVar(&find, "find", false, "Find tools with messages")
	messageCmd.Flags().BoolVar(&list, "list", false, "List available messages for a tool")
	messageCmd.MarkFlagsMutuallyExclusive("find", "list")

	rootCmd.AddCommand(messageCmd)
}

var messageCmd = &cobra.Command{
	Use:       "message",
	Aliases:   []string{"m"},
	Short:     "Show messages for a tool",
	Long:      header + "\nShow messages for a tool",
	Args:      cobra.OnlyValidArgs,
	ValidArgs: tools.GetNames(),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 && !find && !list {
			return nil
		}

		toolName := args[0]

		if list {
			tool, err := tools.GetByName(toolName)
			if err != nil {
				return fmt.Errorf("failed to get tool: %s", err)
			}

			logging.Info.Printfln("Messages for %s:", toolName)
			if tool.Messages.Internals != "" {
				fmt.Println("Internals")
			}
			if tool.Messages.Usage != "" {
				fmt.Println("Usage")
			}
			if tool.Messages.Update != "" {
				fmt.Println("Update")
			}

		} else if find {
			for _, tool := range tools.Tools {
				if tool.Messages.Internals != "" || tool.Messages.Usage != "" || tool.Messages.Update != "" {
					fmt.Println(tool.Name)
				}
			}

		} else {
			printToolInternals(toolName)
			printToolUsage(toolName)
			printToolUpdate(toolName)
			fmt.Println()
		}

		return nil
	},
}

func printToolInternals(toolName string) error {
	return printToolInternalsWithIndentation(toolName, 2)
}

func printToolInternalsWithIndentation(toolName string, indentation int) error {
	tool, err := tools.GetByName(toolName)
	if err != nil {
		return fmt.Errorf("failed to get tool: %s", err)
	}

	if tool.Messages.Internals != "" {
		prefix := pterm.NewStyle(pterm.FgWhite, pterm.BgBlue, pterm.Bold)
		suffix := pterm.NewStyle(pterm.FgWhite)
		prefix.Println()
		prefix.Print(" Internals ")
		suffix.Printfln(" for %s:", tool.Name)
		fmt.Print(tool.ShowInternals(indentation))
	}

	return nil
}

func printToolUsage(toolName string) error {
	return printToolUsageWithIndentation(toolName, 2)
}

func printToolUsageWithIndentation(toolName string, indentation int) error {
	tool, err := tools.GetByName(toolName)
	if err != nil {
		return fmt.Errorf("failed to get tool: %s", err)
	}

	if tool.Messages.Usage != "" {
		prefix := pterm.NewStyle(pterm.FgWhite, pterm.BgGreen, pterm.Bold)
		suffix := pterm.NewStyle(pterm.FgWhite)
		prefix.Println()
		prefix.Print(" Usage ")
		suffix.Printfln(" for %s:", tool.Name)
		fmt.Print(tool.ShowUsage(indentation))
	}

	return nil
}

func printToolUpdate(toolName string) error {
	return printToolUpdateWithIndentation(toolName, 2)
}

func printToolUpdateWithIndentation(toolName string, indentation int) error {
	tool, err := tools.GetByName(toolName)
	if err != nil {
		return fmt.Errorf("failed to get tool: %s", err)
	}

	if tool.Messages.Update != "" {
		prefix := pterm.NewStyle(pterm.FgWhite, pterm.BgYellow, pterm.Bold)
		suffix := pterm.NewStyle(pterm.FgWhite)
		prefix.Println()
		prefix.Print(" Update ")
		suffix.Printfln(" for %s:", tool.Name)
		fmt.Print(tool.ShowUpdate(indentation))
	}

	return nil
}
