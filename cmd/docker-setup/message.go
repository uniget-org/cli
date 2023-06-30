package main

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var find bool
var list bool
var internals bool
var usage bool

func initMessageCmd() {
	messageCmd.Flags().BoolVar(&find, "find", false, "Find tools with messages")
	messageCmd.Flags().BoolVar(&list, "list", false, "List available messages for a tool")
	messageCmd.Flags().BoolVar(&internals, "internals", false, "Show internals for a tool")
	messageCmd.Flags().BoolVar(&usage, "usage", false, "Show usage message for a tool")
	messageCmd.MarkFlagsMutuallyExclusive("find", "list", "usage")

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

		if list {
			tool, err := tools.GetByName(args[0])
			if err != nil {
				return fmt.Errorf("failed to get tool: %s", err)
			}

			pterm.Info.Printfln("Messages for %s:", args[0])
			if tool.Messages.Internals != "" {
				fmt.Println("Internals")
			}
			if tool.Messages.Usage != "" {
				fmt.Println("Usage")
			}

		} else if find {
			for _, tool := range tools.Tools {
				if tool.Messages.Internals != "" || tool.Messages.Usage != "" {
					fmt.Println(tool.Name)
				}
			}

		} else if internals {
			fmt.Println()
			printToolInternals(args[0])
			fmt.Println()

		} else if usage {
			fmt.Println()
			printToolUsage(args[0])
			fmt.Println()

		} else {
			fmt.Println()
			printToolInternals(args[0])
			fmt.Println()
			printToolUsage(args[0])
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

	if tool.Messages.Usage != "" {
		prefix := pterm.NewStyle(pterm.FgWhite, pterm.BgBlue, pterm.Bold)
		suffix := pterm.NewStyle()
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
		suffix := pterm.NewStyle()
		prefix.Print(" Usage ")
		suffix.Printfln(" for %s:", tool.Name)
		fmt.Print(tool.ShowUsage(indentation))
	}

	return nil
}
