package main

import (
	"fmt"
	"html/template"
	"io"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/internal/constants"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/tool"
)

var messageFind bool
var messageList bool

func initMessageCmd() {
	messageCmd.Flags().BoolVar(&messageFind, "find", false, "Find tools with messages")
	messageCmd.Flags().BoolVar(&messageList, "list", false, "List available messages for a tool")
	messageCmd.MarkFlagsMutuallyExclusive("find", "list")

	rootCmd.AddCommand(messageCmd)
}

var messageCmd = &cobra.Command{
	Use: "message",
	Aliases: []string{
		"m",
	},
	Short:   "Show messages for a tool",
	Long:    constants.Header + "\nShow messages for a tool",
	GroupID: "tool",
	Args:    cobra.OnlyValidArgs,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if len(args) == 0 && !messageFind && !messageList {
			return nil
		}

		toolName := args[0]

		if messageList {
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

		} else if messageFind {
			for _, tool := range tools.Tools {

				if tool.Messages.Internals != "" || tool.Messages.Usage != "" || tool.Messages.Update != "" {
					fmt.Println(tool.Name)
				}
			}

		} else {
			err := printToolInternalsMessage(cmd.OutOrStdout(), toolName)
			if err != nil {
				return fmt.Errorf("failed to print tool internals: %s", err)
			}
			err = printToolUsageMessage(cmd.OutOrStdout(), toolName)
			if err != nil {
				return fmt.Errorf("failed to print tool usage: %s", err)
			}
			err = printToolUpdateMessage(cmd.OutOrStdout(), toolName)
			if err != nil {
				return fmt.Errorf("failed to print tool update: %s", err)
			}

			fmt.Println()
		}

		return nil
	},
}

func createTemplateVariablesForTool(tool *tool.Tool) (map[string]any, error) {
	values := make(map[string]any)
	values["Target"] = fmt.Sprintf("%s/%s", configuration.Prefix, configuration.Target)
	values["Name"] = tool.Name
	values["Version"] = tool.Version

	return values, nil
}

func createTemplateVariablesForToolByName(toolName string) (map[string]any, error) {
	tool, err := tools.GetByName(toolName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tool: %s", err)
	}

	return createTemplateVariablesForTool(tool)
}

func printToolInternalsMessage(w io.Writer, toolName string) error {
	values, err := createTemplateVariablesForToolByName(toolName)
	if err != nil {
		return fmt.Errorf("failed to create template variables: %s", err)
	}
	return printToolInternalsMessageWithIndentation(w, toolName, 2, values)
}

func printToolInternalsMessageWithIndentation(w io.Writer, toolName string, indentation int, values map[string]any) error {
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
		output := tool.ShowInternals(indentation)
		tmpl, err := template.New("Internals").Parse(output)
		if err != nil {
			return fmt.Errorf("failed to parse template: %s", err)
		}
		err = tmpl.Execute(w, values)
		if err != nil {
			return fmt.Errorf("failed to execute template: %s", err)
		}
	}

	return nil
}

func printToolUsageMessage(w io.Writer, toolName string) error {
	values, err := createTemplateVariablesForToolByName(toolName)
	if err != nil {
		return fmt.Errorf("failed to create template variables: %s", err)
	}
	return printToolUsageMessageWithIndentation(w, toolName, 2, values)
}

func printToolUsageMessageWithIndentation(w io.Writer, toolName string, indentation int, values map[string]any) error {
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
		output := tool.ShowUsage(indentation)
		tmpl, err := template.New("Internals").Parse(output)
		if err != nil {
			return fmt.Errorf("failed to parse template: %s", err)
		}
		err = tmpl.Execute(w, values)
		if err != nil {
			return fmt.Errorf("failed to execute template: %s", err)
		}
	}

	return nil
}

func printToolUpdateMessage(w io.Writer, toolName string) error {
	values, err := createTemplateVariablesForToolByName(toolName)
	if err != nil {
		return fmt.Errorf("failed to create template variables: %s", err)
	}
	return printToolUpdateMessageWithIndentation(w, toolName, 2, values)
}

func printToolUpdateMessageWithIndentation(w io.Writer, toolName string, indentation int, values map[string]any) error {
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
		output := tool.ShowUpdate(indentation)
		tmpl, err := template.New("Internals").Parse(output)
		if err != nil {
			return fmt.Errorf("failed to parse template: %s", err)
		}
		err = tmpl.Execute(w, values)
		if err != nil {
			return fmt.Errorf("failed to execute template: %s", err)
		}
	}

	return nil
}
