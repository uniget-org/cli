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
		_ = handleError(err, "Failed to mark flag type as required for subcommand hooks add")
	}
	err = addHooksCmd.MarkFlagRequired("source")
	if err != nil {
		_ = handleError(err, "Failed to mark flag source as required for subcommand hooks add")
	}
	hooksCmd.AddCommand(addHooksCmd)

	removeHooksCmd.Flags().StringVar(&hookType, "type", "", "Type of hook to edit (pre-install, post-install, pre-uninstall or post-uninstall)")
	err = removeHooksCmd.MarkFlagRequired("type")
	if err != nil {
		_ = handleError(err, "Failed to mark flag type as required for subcommand hooks remove")
	}
	hooksCmd.AddCommand(removeHooksCmd)

	editHooksCmd.Flags().StringVar(&hookType, "type", "", "Type of hook to edit (pre-install, post-install, pre-uninstall or post-uninstall)")
	err = editHooksCmd.MarkFlagRequired("type")
	if err != nil {
		_ = handleError(err, "Failed to mark flag type as required for subcommand hooks edit")
	}
	hooksCmd.AddCommand(editHooksCmd)

	listHooksCmd.Flags().StringVar(&hookType, "type", "", "Type of hook to list (pre-install, post-install, pre-uninstall or post-uninstall)")
	hooksCmd.AddCommand(listHooksCmd)

	runHooksCmd.Flags().StringVar(&hookType, "type", "", "Type of hook to run (pre-install, post-install, pre-uninstall or post-uninstall)")
	err = runHooksCmd.MarkFlagRequired("type")
	if err != nil {
		_ = handleError(err, "Failed to mark flag type as required for subcommand hooks run")
	}
	hooksCmd.AddCommand(runHooksCmd)

	testHookCmd.Flags().StringVar(&hookType, "type", "", "Type of hook to run (pre-install, post-install, pre-uninstall or post-uninstall)")
	err = testHookCmd.MarkFlagRequired("type")
	if err != nil {
		_ = handleError(err, "Failed to mark flag type as required for subcommand hooks test")
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
	Short:   "Manage hooks",
	Long:    constants.Header + "\nManage hooks\n\nPlease refer to the documentation: https://docs.uniget.dev/hooks/",
	GroupID: "config",
	Args:    cobra.NoArgs,
}

var addHooksCmd = &cobra.Command{
	Use: "add",
	Aliases: []string{
		"a",
	},
	Short: "Add hook",
	Long:  constants.Header + "\nAdd hook",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		sourceInfo, err := os.Stat(hookSource)
		if err != nil {
			return handleError(err, "Hook source file does not exist: %s", hookSource)
		}
		if !sourceInfo.Mode().IsRegular() {
			return handleError(fmt.Errorf("hook source %s is not a regular file", hookSource), "Hook source %s is not a regular file", hookSource)
		}

		hookFileName := filepath.Base(hookSource)
		hookDir := ""
		switch hookType {
		case "pre-install":
			hookDir = configuration.GetHooksPreInstallDirectory()
		case "post-install":
			hookDir = configuration.GetHooksPostInstallDirectory()
		case "pre-uninstall":
			hookDir = configuration.GetHooksPreUninstallDirectory()
		case "post-uninstall":
			hookDir = configuration.GetHooksPostUninstallDirectory()
		default:
			return fmt.Errorf("invalid hook type: %s", hookType)
		}
		myos.AssertDirectory(hookDir)

		hookFile, err := resolveHookFile(hookDir, hookFileName)
		if err != nil {
			return handleError(err, "Unable to resolve hook file %s in %s", hookFileName, hookDir)
		}

		err = myos.CopyFile(hookSource, hookFile)
		if err != nil {
			return handleError(err, "Unable to copy hook file from %s to %s", hookSource, hookFile)
		}

		err = os.Chmod(hookFile, 0700) // #nosec G302 -- File must be executable for execution
		if err != nil {
			return handleError(err, "Unable to set executable permissions on hook file %s", hookFile)
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
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		hookFileName := args[0]
		hookDir := ""
		switch hookType {
		case "pre-install":
			hookDir = configuration.GetHooksPreInstallDirectory()
		case "post-install":
			hookDir = configuration.GetHooksPostInstallDirectory()
		case "pre-uninstall":
			hookDir = configuration.GetHooksPreUninstallDirectory()
		case "post-uninstall":
			hookDir = configuration.GetHooksPostUninstallDirectory()
		default:
			return fmt.Errorf("invalid hook type: %s", hookType)
		}
		myos.AssertDirectory(hookDir)

		hookFile, err := resolveHookFile(hookDir, hookFileName)
		if err != nil {
			return handleError(err, "Hook file %s is outside of directory %s", hookFileName, hookDir)
		}

		if !myos.FileExists(hookFile) {
			return handleError(err, "Hook file does not exist: %s", hookFile)
		}

		err = os.Remove(hookFile)
		if err != nil {
			return handleError(err, "Unable to remove %s hook %s", hookType, hookFileName)
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
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		editorFromVariable := os.Getenv("UNIGET_EDITOR")
		if len(editorFromVariable) == 0 {
			editorFromVariable = os.Getenv("EDITOR")
			if len(editorFromVariable) == 0 {
				return fmt.Errorf("unable to find editor from environment variables UNIGET_EDITOR and EDITOR")
			}
		}
		editor := strings.Split(editorFromVariable, " ")[0]

		hookFileName := args[0]
		hookDir := ""
		switch hookType {
		case "pre-install":
			hookDir = configuration.GetHooksPreInstallDirectory()
		case "post-install":
			hookDir = configuration.GetHooksPostInstallDirectory()
		case "pre-uninstall":
			hookDir = configuration.GetHooksPreUninstallDirectory()
		case "post-uninstall":
			hookDir = configuration.GetHooksPostUninstallDirectory()
		default:
			return fmt.Errorf("invalid hook type: %s", hookType)
		}
		myos.AssertDirectory(hookDir)

		hookFile, err := resolveHookFile(hookDir, hookFileName)
		if err != nil {
			return handleError(err, "Hook file %s is outside of directory %s", hookFileName, hookDir)
		}

		hookFileInfo, err := os.Lstat(hookFile)
		if err != nil {
			return handleError(err, "Unable to stat hook file %s", hookFile)
		}
		if hookFileInfo.Mode()&os.ModeSymlink != 0 {
			return handleError(err, "Hook file %s is a symlink, which is not allowed for security reasons", hookFile)
		}

		command := exec.Command(editor, hookFile) // #nosec G204 -- THis is always the case when relying on EDITOR
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		err = command.Run()
		if err != nil {
			return handleError(err, "Failed to edit %s hook (%s)", hookType, hookFile)
		}

		err = os.Chmod(hookFile, 0700) // #nosec G302 -- File must be executable for execution
		if err != nil {
			return handleError(err, "Unable to set executable permissions on hook file %s", hookFile)
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
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		for _, availableHookType := range []string{"pre-install", "post-install", "pre-uninstall", "post-uninstall"} {
			displayHooks := hookType == "" || availableHookType == hookType

			if displayHooks {
				switch availableHookType {
				case "pre-install":
					preInstallHooksDir := configuration.GetHooksPreInstallDirectory()
					err = processHooks(preInstallHooksDir, func(hookFile string) error {
						fmt.Printf("%s: %s\n", availableHookType, hookFile)
						return nil
					})

				case "post-install":
					postInstallHooksDir := configuration.GetHooksPostInstallDirectory()
					err = processHooks(postInstallHooksDir, func(hookFile string) error {
						fmt.Printf("%s: %s\n", availableHookType, hookFile)
						return nil
					})

				case "pre-uninstall":
					preUninstallHooksDir := configuration.GetHooksPreUninstallDirectory()
					err = processHooks(preUninstallHooksDir, func(hookFile string) error {
						fmt.Printf("%s: %s\n", availableHookType, hookFile)
						return nil
					})

				case "post-uninstall":
					postUninstallHooksDir := configuration.GetHooksPostUninstallDirectory()
					err = processHooks(postUninstallHooksDir, func(hookFile string) error {
						fmt.Printf("%s: %s\n", availableHookType, hookFile)
						return nil
					})
				default:
					return fmt.Errorf("invalid hook type: %s", availableHookType)
				}
				if err != nil {
					return handleError(err, "Unable to list %s hooks", hookType)
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
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		hookDir := ""
		switch hookType {
		case "pre-install":
			hookDir = configuration.GetHooksPreInstallDirectory()
		case "post-install":
			hookDir = configuration.GetHooksPostInstallDirectory()
		case "pre-uninstall":
			hookDir = configuration.GetHooksPreUninstallDirectory()
		case "post-uninstall":
			hookDir = configuration.GetHooksPostUninstallDirectory()
		default:
			return handleError(err, "Invalid hook type: %s", hookType)
		}

		err = processHooks(hookDir, func(hookFile string) error {
			fmt.Printf("Executing %s hook %s:\n", hookType, hookFile)
			_, err := runHook(hookFile, args...)
			return handleError(err, "Unable to execute %s hook %s passing <%v>", hookType, hookFile, args)
		})
		if err != nil {
			return handleError(err, "Unable to execute %s hooks", hookType)
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
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		hookName := args[0]
		hookArgs := args[1:]
		var hookDir string
		switch hookType {
		case "pre-install":
			hookDir = configuration.GetHooksPreInstallDirectory()
		case "post-install":
			hookDir = configuration.GetHooksPostInstallDirectory()
		case "pre-uninstall":
			hookDir = configuration.GetHooksPreUninstallDirectory()
		case "post-uninstall":
			hookDir = configuration.GetHooksPostUninstallDirectory()
		default:
			return handleError(err, "Invalid hook type: %s", hookType)
		}

		hookFile, err := resolveHookFile(hookDir, hookName)
		if err != nil {
			return handleError(err, "Hook file %s is outside of directory %s", hookName, hookDir)
		}

		searchOutputFormat, err := runHook(hookFile, hookArgs...)
		if err != nil {
			return handleError(err, "Unable to execute %s hook %s passing <%v>", hookType, hookName, hookArgs)
		}
		fmt.Print(searchOutputFormat)

		return nil
	},
}

func runPreInstallHooks(args ...string) error {
	if len(args) == 0 {
		return nil
	}

	hookType = "pre-install"
	err := processHooks(configuration.GetHooksPreInstallDirectory(), func(hookFile string) error {
		fmt.Printf("Executing %s hook %s:\n", hookType, hookFile)
		_, err := runHook(hookFile, args...)
		if err != nil {
			return fmt.Errorf("unable to execute %s hook (%s): %s", hookType, hookFile, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to run %s hooks: %s", hookType, err)
	}

	return nil
}

func runPostInstallHooks(args ...string) error {
	if len(args) == 0 {
		return nil
	}

	hookType = "post-install"
	err := processHooks(configuration.GetHooksPostInstallDirectory(), func(hookFile string) error {
		fmt.Printf("Executing %s hook %s:\n", hookType, hookFile)
		_, err := runHook(hookFile, args...)
		if err != nil {
			return fmt.Errorf("unable to execute %s hook (%s): %s", hookType, hookFile, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to run %s hooks: %s", hookType, err)
	}

	return nil
}

func runPreUninstallHooks(args ...string) error {
	if len(args) == 0 {
		return nil
	}

	hookType = "pre-uninstall"
	err := processHooks(configuration.GetHooksPreUninstallDirectory(), func(hookFile string) error {
		fmt.Printf("Executing %s hook %s:\n", hookType, hookFile)
		_, err := runHook(hookFile, args...)
		if err != nil {
			return fmt.Errorf("unable to execute %s hook (%s): %s", hookType, hookFile, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to run %s hooks: %s", hookType, err)
	}

	return nil
}

func runPostUninstallHooks(args ...string) error {
	if len(args) == 0 {
		return nil
	}

	hookType = "post-uninstall"
	err := processHooks(configuration.GetHooksPostUninstallDirectory(), func(hookFile string) error {
		fmt.Printf("Executing %s hook %s:\n", hookType, hookFile)
		_, err := runHook(hookFile, args...)
		if err != nil {
			return fmt.Errorf("unable to execute %s hook (%s): %s", hookType, hookFile, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to run %s hooks: %s", hookType, err)
	}

	return nil
}

func processHooks(path string, callback func(file string) error) error {
	if !myos.DirectoryExists(path) {
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

		logging.Debugf("processing hook file %s/%s", path, file.Name())

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
	absHookFile, err := filepath.Abs(hookFile)
	if err != nil {
		return "", fmt.Errorf("unable to get absolute path of hook file %s: %w", hookFile, err)
	}
	absHooksDir, err := filepath.Abs(configuration.GetHooksRootDirectory())
	if err != nil {
		return "", fmt.Errorf("unable to get absolute path of hooks directory: %w", err)
	}
	if !myos.PathIsWithin(absHooksDir, absHookFile) {
		return "", fmt.Errorf("hook file %s is not inside the allowed hooks directory %s", hookFile, absHooksDir)
	}

	if !myos.FileExists(hookFile) {
		return "", fmt.Errorf("hook does not exist: %s", hookFile)
	}

	logging.Debugf("running hook in file %s (args: %s)", hookFile, args)
	command := exec.Command(hookFile, args...) // #nosec G204 -- Tool images are a trusted source
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err = command.Run()
	if err != nil {
		fmt.Printf("stdout:\n%s\n", command.Stdout)
		fmt.Printf("stderr:\n%s\n", command.Stderr)
		return "", fmt.Errorf("unable to execute hook (%s): %s", hookFile, err)
	}

	return string(searchOutputFormat), nil
}

// resolveHookFile resolves name relative to directory and verifies that the
// result does not escape directory.
func resolveHookFile(directory, name string) (string, error) {
	if len(name) == 0 {
		return "", fmt.Errorf("hook file name must not be empty")
	}
	if name == "." || !filepath.IsLocal(name) {
		return "", fmt.Errorf("hook file name %s is not a safe path component", name)
	}
	if filepath.IsAbs(name) {
		return "", fmt.Errorf("hook file name %s must be relative", name)
	}

	absDirectory, err := filepath.Abs(directory)
	if err != nil {
		return "", fmt.Errorf("unable to get absolute path of directory %s: %w", directory, err)
	}

	absFile, err := filepath.Abs(filepath.Join(absDirectory, name))
	if err != nil {
		return "", fmt.Errorf("unable to get absolute path of hook file %s: %w", name, err)
	}

	if !myos.PathIsWithin(absDirectory, absFile) {
		return "", fmt.Errorf("hook file %s is outside of directory %s", name, directory)
	}

	return absFile, nil
}
