package main

import (
	"fmt"
	"html/template"
	"io"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uniget-org/cli/pkg/logging"
	"github.com/uniget-org/cli/pkg/tool"
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
	Use:     "message",
	Aliases: []string{"m"},
	Short:   "Show messages for a tool",
	Long:    header + "\nShow messages for a tool",
	Args:    cobra.OnlyValidArgs,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
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
			checkClientVersionRequirement(tool)

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
			for index, tool := range tools.Tools {
				checkClientVersionRequirement(&tools.Tools[index])

				if tool.Messages.Internals != "" || tool.Messages.Usage != "" || tool.Messages.Update != "" {
					fmt.Println(tool.Name)
				}
			}

		} else {
			err := printToolInternals(cmd.OutOrStdout(), toolName)
			if err != nil {
				return fmt.Errorf("failed to print tool internals: %s", err)
			}
			err = printToolUsage(cmd.OutOrStdout(), toolName)
			if err != nil {
				return fmt.Errorf("failed to print tool usage: %s", err)
			}
			err = printToolUpdate(cmd.OutOrStdout(), toolName)
			if err != nil {
				return fmt.Errorf("failed to print tool update: %s", err)
			}

			fmt.Println()
		}

		return nil
	},
}

func createTemplateVariablesForTool(tool *tool.Tool) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	values["Target"] = fmt.Sprintf("%s/%s", viper.GetString("prefix"), viper.GetString("target"))
	values["Name"] = tool.Name
	values["Version"] = tool.Version

	return values, nil
}

func createTemplateVariablesForToolByName(toolName string) (map[string]interface{}, error) {
	tool, err := tools.GetByName(toolName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tool: %s", err)
	}

	return createTemplateVariablesForTool(tool)
}

func printToolInternals(w io.Writer, toolName string) error {
	values, err := createTemplateVariablesForToolByName(toolName)
	if err != nil {
		return fmt.Errorf("failed to create template variables: %s", err)
	}
	return printToolInternalsWithIndentation(w, toolName, 2, values)
}

func printToolInternalsWithIndentation(w io.Writer, toolName string, indentation int, values map[string]interface{}) error {
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

func printToolUsage(w io.Writer, toolName string) error {
	values, err := createTemplateVariablesForToolByName(toolName)
	if err != nil {
		return fmt.Errorf("failed to create template variables: %s", err)
	}
	return printToolUsageWithIndentation(w, toolName, 2, values)
}

func printToolUsageWithIndentation(w io.Writer, toolName string, indentation int, values map[string]interface{}) error {
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

func printToolUpdate(w io.Writer, toolName string) error {
	values, err := createTemplateVariablesForToolByName(toolName)
	if err != nil {
		return fmt.Errorf("failed to create template variables: %s", err)
	}
	return printToolUpdateWithIndentation(w, toolName, 2, values)
}

func printToolUpdateWithIndentation(w io.Writer, toolName string, indentation int, values map[string]interface{}) error {
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
