package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gitlab.com/uniget-org/cli/pkg/logging"
	myos "gitlab.com/uniget-org/cli/pkg/os"
)

var (
	hookType   = ""
	hookSource = ""
)

func initHooksCmd() {
	var err error

	addHooksCmd.Flags().StringVar(&hookType, "type", "", "Type of hook to add (pre or post)")
	addHooksCmd.Flags().StringVar(&hookSource, "source", "", "Path to the hook script")
	err = addHooksCmd.MarkFlagRequired("type")
	if err != nil {
		logging.Error.Printfln("Failed to mark flag as required: %v", err)
	}
	err = addHooksCmd.MarkFlagRequired("source")
	if err != nil {
		logging.Error.Printfln("Failed to mark flag as required: %v", err)
	}
	hooksCmd.AddCommand(addHooksCmd)

	runHooksCmd.Flags().StringVar(&hookType, "type", "", "Type of hook to run (pre or post)")
	err = runHooksCmd.MarkFlagRequired("type")
	if err != nil {
		logging.Error.Printfln("Failed to mark flag as required: %v", err)
	}
	hooksCmd.AddCommand(runHooksCmd)

	editHooksCmd.Flags().StringVar(&hookType, "type", "", "Type of hook to run (pre or post)")
	err = editHooksCmd.MarkFlagRequired("type")
	if err != nil {
		logging.Error.Printfln("Failed to mark flag as required: %v", err)
	}
	hooksCmd.AddCommand(editHooksCmd)

	rootCmd.AddCommand(hooksCmd)
}

var hooksCmd = &cobra.Command{
	Use: "hooks",
	Aliases: []string{
		"hook",
		"h",
	},
	Short:  "Manage hooks",
	Long:   header + "\nManage hooks",
	Hidden: true,
	Args:   cobra.NoArgs,
}

var addHooksCmd = &cobra.Command{
	Use: "add",
	Aliases: []string{
		"a",
	},
	Short: "Add hook",
	Long:  header + "\nAdd hook",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !fileExists(hookSource) {
			return fmt.Errorf("hook source file does not exist: %s", hookSource)
		}

		preHooksDir := viper.GetString("prefix") + "/" + configDirectory + "/" + hooksPreDirectory
		postHooksDir := viper.GetString("prefix") + "/" + configDirectory + "/" + hooksPostDirectory

		hookSourceSplit := strings.Split(hookSource, "/")
		hookFileName := hookSourceSplit[len(hookSourceSplit)-1]
		hookFile := ""
		switch hookType {
		case "pre":
			assertDirectory(preHooksDir)
			hookFile = preHooksDir + "/" + hookFileName
		case "post":
			assertDirectory(postHooksDir)
			hookFile = postHooksDir + "/" + hookFileName
		}

		err := myos.CopyFile(hookSource, hookFile)
		if err != nil {
			return fmt.Errorf("unable to copy hook file from %s to %s: %w", hookSource, hookFile, err)
		}

		err = os.Chmod(hookFile, 0700) // #nosec G302 -- File must be executable for execution
		if err != nil {
			return fmt.Errorf("unable to set executable permissions on hook file %s: %w", hookFile, err)
		}

		return nil
	},
}

var editHooksCmd = &cobra.Command{
	Use: "edit",
	Aliases: []string{
		"ed",
		"e",
	},
	Short: "Edit hook",
	Long:  header + "\nEdit hook",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		preHooksDir := viper.GetString("prefix") + "/" + configDirectory + "/" + hooksPreDirectory
		postHooksDir := viper.GetString("prefix") + "/" + configDirectory + "/" + hooksPostDirectory

		hookFileName := args[0]
		hookFile := ""
		switch hookType {
		case "pre":
			assertDirectory(preHooksDir)
			hookFile = preHooksDir + "/" + hookFileName
		case "post":
			assertDirectory(postHooksDir)
			hookFile = postHooksDir + "/" + hookFileName
		}

		command := exec.Command("vim", hookFile)
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		err = command.Run()
		if err != nil {
			return fmt.Errorf("failed to edit %s hook (%s): %s", hookType, hookFile, err)
		}

		err = os.Chmod(hookFile, 0700) // #nosec G302 -- File must be executable for execution
		if err != nil {
			return fmt.Errorf("unable to set executable permissions on hook file %s: %w", hookFile, err)
		}

		return nil
	},
}

var runHooksCmd = &cobra.Command{
	Use: "run",
	Aliases: []string{
		"r",
	},
	Short: "Run hooks",
	Long:  header + "\nRun hooks",
	Args:  cobra.MinimumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		preHooksDir := viper.GetString("prefix") + "/" + configDirectory + "/" + hooksPreDirectory
		postHooksDir := viper.GetString("prefix") + "/" + configDirectory + "/" + hooksPostDirectory

		hookFile := ""
		switch hookType {
		case "pre":
			hookFile = preHooksDir + "/foo.sh"
		case "post":
			hookFile = postHooksDir + "/foo.sh"
		}

		command := exec.Command(hookFile, args...) // #nosec G204 -- Tool images are a trusted source
		_, err := command.Output()
		if err != nil {
			return fmt.Errorf("unable to execute %s hook (%s): %s", hookType, hookFile, err)
		}

		return nil
	},
}
