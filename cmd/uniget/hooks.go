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

	addHooksCmd.Flags().StringVar(&hookType, "type", "", "Type of hook to add (pre-install, post-install, pre-uninstall or post-uninstall)")
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

	removeHooksCmd.Flags().StringVar(&hookType, "type", "", "Type of hook to edit (pre-install, post-install, pre-uninstall or post-uninstall)")
	err = removeHooksCmd.MarkFlagRequired("type")
	if err != nil {
		logging.Error.Printfln("Failed to mark flag as required: %v", err)
	}
	hooksCmd.AddCommand(removeHooksCmd)

	editHooksCmd.Flags().StringVar(&hookType, "type", "", "Type of hook to edit (pre-install, post-install, pre-uninstall or post-uninstall)")
	err = editHooksCmd.MarkFlagRequired("type")
	if err != nil {
		logging.Error.Printfln("Failed to mark flag as required: %v", err)
	}
	hooksCmd.AddCommand(editHooksCmd)

	listHooksCmd.Flags().StringVar(&hookType, "type", "", "Type of hook to list (pre-install, post-install, pre-uninstall or post-uninstall)")
	hooksCmd.AddCommand(listHooksCmd)

	runHooksCmd.Flags().StringVar(&hookType, "type", "", "Type of hook to run (pre-install, post-install, pre-uninstall or post-uninstall)")
	err = runHooksCmd.MarkFlagRequired("type")
	if err != nil {
		logging.Error.Printfln("Failed to mark flag as required: %v", err)
	}
	hooksCmd.AddCommand(runHooksCmd)

	testHookCmd.Flags().StringVar(&hookType, "type", "", "Type of hook to run (pre-install, post-install, pre-uninstall or post-uninstall)")
	err = testHookCmd.MarkFlagRequired("type")
	if err != nil {
		logging.Error.Printfln("Failed to mark flag as required: %v", err)
	}
	hooksCmd.AddCommand(testHookCmd)

	rootCmd.AddCommand(hooksCmd)
}

var hooksCmd = &cobra.Command{
	Use: "hooks",
	Aliases: []string{
		"hook",
		"h",
	},
	Short: "Manage hooks",
	Long:  header + "\nManage hooks",
	Args:  cobra.NoArgs,
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

		hooksDir := viper.GetString("prefix") + "/" + configDirectory

		hookSourceSplit := strings.Split(hookSource, "/")
		hookFileName := hookSourceSplit[len(hookSourceSplit)-1]
		hookFile := ""
		switch hookType {
		case "pre-install":
			preInstallHooksDir := hooksDir + "/" + hooksPreInstallDirectory
			assertDirectory(preInstallHooksDir)
			hookFile = preInstallHooksDir + "/" + hookFileName
		case "post-install":
			postInstallHooksDir := hooksDir + "/" + hooksPostInstallDirectory
			assertDirectory(postInstallHooksDir)
			hookFile = postInstallHooksDir + "/" + hookFileName
		case "pre-uninstall":
			preUninstallHooksDir := hooksDir + "/" + hooksPreInstallDirectory
			assertDirectory(preUninstallHooksDir)
			hookFile = preUninstallHooksDir + "/" + hookFileName
		case "post-uninstall":
			postUninstallHooksDir := hooksDir + "/" + hooksPostInstallDirectory
			assertDirectory(postUninstallHooksDir)
			hookFile = postUninstallHooksDir + "/" + hookFileName
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

var removeHooksCmd = &cobra.Command{
	Use: "remove",
	Aliases: []string{
		"r",
		"rm",
		"delete",
		"del",
		"d",
	},
	Short: "Add hook",
	Long:  header + "\nAdd hook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		hookFileName := args[0]
		hooksDir := viper.GetString("prefix") + "/" + configDirectory
		hookFile := ""
		switch hookType {
		case "pre-install":
			preInstallHooksDir := hooksDir + "/" + hooksPreInstallDirectory
			hookFile = preInstallHooksDir + "/" + hookFileName
		case "post-install":
			postInstallHooksDir := hooksDir + "/" + hooksPostInstallDirectory
			assertDirectory(postInstallHooksDir)
			hookFile = postInstallHooksDir + "/" + hookFileName
		case "pre-uninstall":
			preUninstallHooksDir := hooksDir + "/" + hooksPreInstallDirectory
			assertDirectory(preUninstallHooksDir)
			hookFile = preUninstallHooksDir + "/" + hookFileName
		case "post-uninstall":
			postUninstallHooksDir := hooksDir + "/" + hooksPostInstallDirectory
			assertDirectory(postUninstallHooksDir)
			hookFile = postUninstallHooksDir + "/" + hookFileName
		}

		if !fileExists(hookFile) {
			return fmt.Errorf("hook file does not exist: %s", hookFile)
		}

		err = os.Remove(hookFile)
		if err != nil {
			return fmt.Errorf("unable to remove %s hook %s: %s", hookType, hookFileName, err)
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
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		editor := os.Getenv("EDITOR")
		if len(editor) == 0 {
			return fmt.Errorf("unable to find editor from environment")
		}

		hooksDir := viper.GetString("prefix") + "/" + configDirectory

		hookFileName := args[0]
		hookDir := ""
		hookFile := ""
		switch hookType {
		case "pre-install":
			preInstallHooksDir := hooksDir + "/" + hooksPreInstallDirectory
			hookDir = preInstallHooksDir
			hookFile = hookDir + "/" + hookFileName
		case "post-install":
			postInstallHooksDir := hooksDir + "/" + hooksPostInstallDirectory
			hookDir = postInstallHooksDir
			hookFile = hookDir + "/" + hookFileName
		case "pre-uninstall":
			preUninstallHooksDir := hooksDir + "/" + hooksPreInstallDirectory
			hookDir = preUninstallHooksDir
			hookFile = hookDir + "/" + hookFileName
		case "post-uninstall":
			postUninstallHooksDir := hooksDir + "/" + hooksPostInstallDirectory
			hookDir = postUninstallHooksDir
			hookFile = hookDir + "/" + hookFileName
		}
		assertDirectory(hookDir)

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

var listHooksCmd = &cobra.Command{
	Use: "list",
	Aliases: []string{
		"l",
		"show",
	},
	Short: "List hooks",
	Long:  header + "\nList hooks",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		hooksDir := viper.GetString("prefix") + "/" + configDirectory

		for _, availableHookType := range []string{"pre-install", "post-install", "pre-uninstall", "post-uninstall"} {
			displayHooks := hookType == "" || availableHookType == hookType

			if displayHooks {
				switch availableHookType {
				case "pre-install":
					preInstallHooksDir := hooksDir + "/" + hooksPreInstallDirectory
					err = processHooks(preInstallHooksDir, func(hookFile string) error {
						fmt.Printf("%s: %s\n", availableHookType, hookFile)
						return nil
					})

				case "post-install":
					postInstallHooksDir := hooksDir + "/" + hooksPostInstallDirectory
					err = processHooks(postInstallHooksDir, func(hookFile string) error {
						fmt.Printf("%s: %s\n", availableHookType, hookFile)
						return nil
					})

				case "pre-uninstall":
					preUninstallHooksDir := hooksDir + "/" + hooksPreUninstallDirectory
					err = processHooks(preUninstallHooksDir, func(hookFile string) error {
						fmt.Printf("%s: %s\n", availableHookType, hookFile)
						return nil
					})

				case "post-uninstall":
					postUninstallHooksDir := hooksDir + "/" + hooksPostUninstallDirectory
					err = processHooks(postUninstallHooksDir, func(hookFile string) error {
						fmt.Printf("%s: %s\n", availableHookType, hookFile)
						return nil
					})
				}
				if err != nil {
					return fmt.Errorf("unable to list %s hooks: %s", hookType, err)
				}
			}
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
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		switch hookType {
		case "pre-install":
			err = runHooks(hooksPreInstallDirectory, args...)
		case "post-install":
			err = runHooks(hooksPostInstallDirectory, args...)
		case "pre-uninstall":
			err = runHooks(hooksPreUninstallDirectory, args...)
		case "post-uninstall":
			err = runHooks(hooksPostUninstallDirectory, args...)
		}
		if err != nil {
			return fmt.Errorf("unable to execute %s hooks: %s", hookType, err)
		}

		return nil
	},
}

var testHookCmd = &cobra.Command{
	Use: "test",
	Aliases: []string{
		"t",
	},
	Short: "Test single hook",
	Long:  header + "\nTest single hook",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		hooksDir := viper.GetString("prefix") + "/" + configDirectory
		hookName := args[0]
		hookArgs := args[1:]
		var hookFile string
		switch hookType {
		case "pre-install":
			hookFile = hooksDir + "/" + hooksPreInstallDirectory + "/" + hookName
		case "post-install":
			hookFile = hooksDir + "/" + hooksPostInstallDirectory + "/" + hookName
		case "pre-uninstall":
			hookFile = hooksDir + "/" + hooksPreUninstallDirectory + "/" + hookName
		case "post-uninstall":
			hookFile = hooksDir + "/" + hooksPostUninstallDirectory + "/" + hookName
		}

		output, err = runHook(hookFile, hookArgs...)
		if err != nil {
			return fmt.Errorf("unable to execute %s hook %s passing <%v>: %s", hookType, hookName, hookArgs, err)
		}
		fmt.Print(output)

		return nil
	},
}

func runHooks(hookTypePath string, args ...string) error {
	hooksDir := viper.GetString("prefix") + "/" + configDirectory + "/" + hookTypePath
	err := processHooks(hooksDir, func(hookFile string) error {
		_, err := runHook(hookFile, args...)
		return err
	})
	if err != nil {
		return fmt.Errorf("unable to run pre hooks: %s", err)
	}

	return nil
}

func runPreInstallHooks(args ...string) error {
	return runHooks(hooksPreInstallDirectory, args...)
}

func runPostInstallHooks(args ...string) error {
	return runHooks(hooksPostInstallDirectory, args...)
}

func runPreUninstallHooks(args ...string) error {
	return runHooks(hooksPreUninstallDirectory, args...)
}

func runPostUninstallHooks(args ...string) error {
	return runHooks(hooksPostUninstallDirectory, args...)
}

func processHooks(path string, callback func(file string) error) error {
	if !directoryExists(path) {
		return nil
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("unable to read directory %s: %w", path, err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		hookFile := path + "/" + file.Name()
		err := callback(hookFile)
		if err != nil {
			return fmt.Errorf("error processing hook file %s: %w", hookFile, err)
		}
	}

	return nil
}

func runHook(hookFile string, args ...string) (string, error) {
	if !fileExists(hookFile) {
		return "", fmt.Errorf("hook does not exist: %s", hookFile)
	}

	logging.Debugf("running hook in file %s (args: %s)", hookFile, args)
	command := exec.Command(hookFile, args...) // #nosec G204 -- Tool images are a trusted source
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("unable to execute %s hook (%s): %s", hookType, hookFile, err)
	}

	return string(output), nil
}
