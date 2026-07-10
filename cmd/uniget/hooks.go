package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"gitlab.com/uniget-org/cli/internal/constants"
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
	Long:  constants.Header + "\nManage hooks\n\nPlease refer to the documentation: https://docs.uniget.dev/hooks/",
	Args:  cobra.NoArgs,
}

var addHooksCmd = &cobra.Command{
	Use: "add",
	Aliases: []string{
		"a",
	},
	Short: "Add hook",
	Long:  constants.Header + "\nAdd hook",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !fileExists(hookSource) {
			return fmt.Errorf("hook source file does not exist: %s", hookSource)
		}

		hooksDir := configuration.Prefix + "/" + configuration.GetConfigDirectory()

		hookSourceSplit := strings.Split(hookSource, "/")
		hookFileName := hookSourceSplit[len(hookSourceSplit)-1]
		hookFile := ""
		switch hookType {
		case "pre-install":
			preInstallHooksDir := hooksDir + "/" + constants.HooksPreInstallDirectory
			assertDirectory(preInstallHooksDir)
			hookFile = preInstallHooksDir + "/" + hookFileName
		case "post-install":
			postInstallHooksDir := hooksDir + "/" + constants.HooksPostInstallDirectory
			assertDirectory(postInstallHooksDir)
			hookFile = postInstallHooksDir + "/" + hookFileName
		case "pre-uninstall":
			preUninstallHooksDir := hooksDir + "/" + constants.HooksPreUninstallDirectory
			assertDirectory(preUninstallHooksDir)
			hookFile = preUninstallHooksDir + "/" + hookFileName
		case "post-uninstall":
			postUninstallHooksDir := hooksDir + "/" + constants.HooksPostUninstallDirectory
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
	Short: "Remove hook",
	Long:  constants.Header + "\nRemove hook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		hookFileName := args[0]
		hooksDir := configuration.Prefix + "/" + configuration.GetConfigDirectory()
		hookFile := ""
		switch hookType {
		case "pre-install":
			preInstallHooksDir := hooksDir + "/" + constants.HooksPreInstallDirectory
			hookFile = preInstallHooksDir + "/" + hookFileName
		case "post-install":
			postInstallHooksDir := hooksDir + "/" + constants.HooksPostInstallDirectory
			assertDirectory(postInstallHooksDir)
			hookFile = postInstallHooksDir + "/" + hookFileName
		case "pre-uninstall":
			preUninstallHooksDir := hooksDir + "/" + constants.HooksPreUninstallDirectory
			assertDirectory(preUninstallHooksDir)
			hookFile = preUninstallHooksDir + "/" + hookFileName
		case "post-uninstall":
			postUninstallHooksDir := hooksDir + "/" + constants.HooksPostUninstallDirectory
			assertDirectory(postUninstallHooksDir)
			hookFile = postUninstallHooksDir + "/" + hookFileName
		}

		hookFileAbs, err := filepath.Abs(hookFile)
		if err != nil {
			return fmt.Errorf("unable to get absolute path of hook file %s: %w", hookFile, err)
		}
		if !strings.HasPrefix(hookFileAbs, hooksDir) {
			return fmt.Errorf("hook file %s is outside of hookDir %s", hookFile, hooksDir)
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
	Long:  constants.Header + "\nEdit hook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		editorFromVariable := os.Getenv("UNIGET_EDITOR")
		if len(editorFromVariable) == 0 {
			editorFromVariable = os.Getenv("EDITOR")
			if len(editorFromVariable) == 0 {
				return fmt.Errorf("unable to find editor from environment variables UNIGET_EDITOR and EDITOR")
			}
		}
		editor := strings.Split(editorFromVariable, " ")[0]

		hooksDir := configuration.Prefix + "/" + configuration.GetConfigDirectory()

		hookFileName := args[0]
		hookDir := ""
		hookFile := ""
		switch hookType {
		case "pre-install":
			preInstallHooksDir := hooksDir + "/" + constants.HooksPreInstallDirectory
			hookDir = preInstallHooksDir
			hookFile = hookDir + "/" + hookFileName
		case "post-install":
			postInstallHooksDir := hooksDir + "/" + constants.HooksPostInstallDirectory
			hookDir = postInstallHooksDir
			hookFile = hookDir + "/" + hookFileName
		case "pre-uninstall":
			preUninstallHooksDir := hooksDir + "/" + constants.HooksPreUninstallDirectory
			hookDir = preUninstallHooksDir
			hookFile = hookDir + "/" + hookFileName
		case "post-uninstall":
			postUninstallHooksDir := hooksDir + "/" + constants.HooksPostUninstallDirectory
			hookDir = postUninstallHooksDir
			hookFile = hookDir + "/" + hookFileName
		}
		assertDirectory(hookDir)

		hookFileAbs, err := filepath.Abs(hookFile)
		if err != nil {
			return fmt.Errorf("unable to get absolute path of hook file %s: %w", hookFile, err)
		}
		if !strings.HasPrefix(hookFileAbs, hooksDir) {
			return fmt.Errorf("hook file %s is outside of hookDir %s", hookFile, hooksDir)
		}

		hookFileInfo, err := os.Lstat(hookFile)
		if err != nil {
			return fmt.Errorf("unable to stat hook file %s: %w", hookFile, err)
		}
		if hookFileInfo.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("hook file %s is a symlink, which is not allowed for security reasons", hookFile)
		}

		command := exec.Command(editor, hookFile) // #nosec G204 -- THis is always the case when relying on EDITOR
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
	Long:  constants.Header + "\nList hooks",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		hooksDir := configuration.Prefix + "/" + configuration.GetConfigDirectory()

		for _, availableHookType := range []string{"pre-install", "post-install", "pre-uninstall", "post-uninstall"} {
			displayHooks := hookType == "" || availableHookType == hookType

			if displayHooks {
				switch availableHookType {
				case "pre-install":
					preInstallHooksDir := hooksDir + "/" + constants.HooksPreInstallDirectory
					err = processHooks(preInstallHooksDir, func(hookFile string) error {
						fmt.Printf("%s: %s\n", availableHookType, hookFile)
						return nil
					})

				case "post-install":
					postInstallHooksDir := hooksDir + "/" + constants.HooksPostInstallDirectory
					err = processHooks(postInstallHooksDir, func(hookFile string) error {
						fmt.Printf("%s: %s\n", availableHookType, hookFile)
						return nil
					})

				case "pre-uninstall":
					preUninstallHooksDir := hooksDir + "/" + constants.HooksPreUninstallDirectory
					err = processHooks(preUninstallHooksDir, func(hookFile string) error {
						fmt.Printf("%s: %s\n", availableHookType, hookFile)
						return nil
					})

				case "post-uninstall":
					postUninstallHooksDir := hooksDir + "/" + constants.HooksPostUninstallDirectory
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
	Long:  constants.Header + "\nRun hooks",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		switch hookType {
		case "pre-install":
			err = runHooks(hookType, constants.HooksPreInstallDirectory, args...)
		case "post-install":
			err = runHooks(hookType, constants.HooksPostInstallDirectory, args...)
		case "pre-uninstall":
			err = runHooks(hookType, constants.HooksPreUninstallDirectory, args...)
		case "post-uninstall":
			err = runHooks(hookType, constants.HooksPostUninstallDirectory, args...)
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
	Long:  constants.Header + "\nTest single hook",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		hooksDir := configuration.Prefix + "/" + configuration.GetConfigDirectory()
		hookName := args[0]
		hookArgs := args[1:]
		var hookFile string
		switch hookType {
		case "pre-install":
			hookFile = hooksDir + "/" + constants.HooksPreInstallDirectory + "/" + hookName
		case "post-install":
			hookFile = hooksDir + "/" + constants.HooksPostInstallDirectory + "/" + hookName
		case "pre-uninstall":
			hookFile = hooksDir + "/" + constants.HooksPreUninstallDirectory + "/" + hookName
		case "post-uninstall":
			hookFile = hooksDir + "/" + constants.HooksPostUninstallDirectory + "/" + hookName
		}

		output, err = runHook(hookFile, hookArgs...)
		if err != nil {
			return fmt.Errorf("unable to execute %s hook %s passing <%v>: %s", hookType, hookName, hookArgs, err)
		}
		fmt.Print(output)

		return nil
	},
}

func runHooks(hookType string, hookTypePath string, args ...string) error {
	hooksDir := configuration.Prefix + "/" + configuration.GetConfigDirectory() + "/" + hookTypePath
	err := processHooks(hooksDir, func(hookFile string) error {
		fmt.Printf("Executing %s hook %s:\n", hookType, hookFile)
		_, err := runHook(hookFile, args...)
		return err
	})
	if err != nil {
		return fmt.Errorf("unable to run pre hooks: %s", err)
	}

	return nil
}

func runPreInstallHooks(args ...string) error {
	if len(args) == 0 {
		return nil
	}
	return runHooks("pre-install", constants.HooksPreInstallDirectory, args...)
}

func runPostInstallHooks(args ...string) error {
	if len(args) == 0 {
		return nil
	}
	return runHooks("post-install", constants.HooksPostInstallDirectory, args...)
}

func runPreUninstallHooks(args ...string) error {
	if len(args) == 0 {
		return nil
	}
	return runHooks("pre-uninstall", constants.HooksPreUninstallDirectory, args...)
}

func runPostUninstallHooks(args ...string) error {
	if len(args) == 0 {
		return nil
	}
	return runHooks("post-uninstall", constants.HooksPostUninstallDirectory, args...)
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
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}

		hookFile := path + "/" + file.Name()
		hookFileInfo, err := os.Lstat(hookFile)
		if err != nil {
			return fmt.Errorf("unable to stat hook file %s: %w", hookFile, err)
		}
		if hookFileInfo.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("hook file %s is a symlink, which is not allowed for security reasons", hookFile)
		}

		err = callback(hookFile)
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
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err := command.Run()
	if err != nil {
		return "", fmt.Errorf("unable to execute %s hook (%s): %s", hookType, hookFile, err)
	}

	return string(output), nil
}
