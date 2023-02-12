package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"
	//"github.com/fatih/color"

	"github.com/nicholasdille/docker-setup/pkg/tool"
	//"github.com/nicholasdille/docker-setup/pkg/shell"
)

var installMode string
var check bool
var plan bool
var toolStatus map[string]tool.ToolStatus = make(map[string]tool.ToolStatus)
var requestedTools tool.Tools
var plannedTools tool.Tools
var no_wait bool

//var check_mark string = "✓" // Unicode=\u2713 UTF-8=\xE2\x9C\x93 (https://www.compart.com/de/unicode/U+2713)
//var cross_mark string = "✗" // Unicode=\u2717 UTF-8=\xE2\x9C\x97 (https://www.compart.com/de/unicode/U+2717)

func initInstallCmd() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().StringVarP(&installMode, "mode",    "m", "default", "How to install (default, list, tags, installed)")
	installCmd.Flags().BoolVarP(  &plan,        "plan",    "p", false,     "Show planned installations")
	installCmd.Flags().BoolVarP(  &check,       "check",   "c", false,     "Abort after checking versions")
	installCmd.Flags().BoolVarP(  &no_wait,     "no-wait", "n", false,     "Skip wait before installation")

	installCmd.Flags().BoolP("skip-docs",       "s", false, "Do not install documentation for faster installation")
	installCmd.Flags().BoolP("reinstall",       "r", false, "Reinstall tools")
	installCmd.Flags().BoolP("no-cache",        "",  false, "Do not cache downloads")
	installCmd.Flags().BoolP("no-cron",         "",  false, "Do not create cronjob for automated updates")
}

var installCmd = &cobra.Command{
	Use:       "install [tool...]",
	Aliases:   []string{"i"},
	Short:     "Install tools",
	Long:      header + "\nInstall and update tools",
	ValidArgs: tools.GetNames(),
	Args:      cobra.OnlyValidArgs,
	RunE:      func(cmd *cobra.Command, args []string) error {

		// Validation checks
		log.Tracef("Found %d argument(s): %+v", len(args), args)
		if installMode == "list" || installMode == "tags" {
			if len(args) == 0 {
				return fmt.Errorf("You must specify at least one tool for mode list or tags.")
			}
		}
		if check && plan {
			return fmt.Errorf("You can only only specify one: --check, --plan")
		}

		// Fill default values and replace variables
		for index, tool := range tools.Tools {
			log.Tracef("Getting status for requested tool %s", tool.Name)
			tools.Tools[index].ReplaceVariables(target, arch, alt_arch)

			status, err := tools.Tools[index].GetStatus()
			if err != nil {
				return fmt.Errorf("Unable to determine status of %s: %s", tool.Name, err)
			}
			
			toolStatus[tool.Name] = status
		}

		// Collect requested tools based on mode
		if installMode == "list" {
			requestedTools = tools.GetByNames(args)

		} else if installMode == "tags" {
			requestedTools = tools.GetByTags(args)

		} else if installMode == "default" {
			requestedTools = tools

		} else if installMode == "only-installed" {
			for _, tool := range tools.Tools {
				if toolStatus[tool.Name].BinaryPresent {
					requestedTools.Tools = append(requestedTools.Tools, tool)
				}
			}
		}
		log.Debugf("Requested %d tool(s)", len(requestedTools.Tools))

		// Add dependencies of requested tools
		// Set installation order
		for _, tool := range requestedTools.Tools {
			err := tools.ResolveDependencies(&plannedTools, tool.Name)
			if err != nil {
				return fmt.Errorf("Unable to resolve dependencies for %s: %s", tool.Name, err)
			}
		}
		log.Debugf("Planned %d tool(s)", len(plannedTools.Tools))

		// Terminate if checking or planning
		if plan || check {
			plannedTools.ListWithStatus(toolStatus)
		}
		if check {
			for _, tool := range plannedTools.Tools {
				if ! toolStatus[tool.Name].BinaryPresent || ! toolStatus[tool.Name].VersionMatches {
					return fmt.Errorf("Found missing or outdated tool")
				}
			}
		}
		if plan || check {
			return nil
		}

		// Wait before installation
		if ! no_wait {
			log.Info("Press Ctrl-C to interrupt...")
			time.Sleep(10 * time.Second)
		}

		// Install
		for _, tool := range plannedTools.Tools {
			log.Infof("Installing %s", tool.Name)
			err := tool.Install(alt_arch)
			if err != nil {
				return fmt.Errorf("Unable to install downloads: %s", err)
			}
		}

		return nil

		//toolName := "docker"
		//toolCacheDirectory := cacheDirectory + "/" + toolName
		//toolInstallScript := toolCacheDirectory + "/install.sh"
		//os.MkdirAll(toolCacheDirectory, 0755)
		//err := shell.CreateScript(toolInstallScript, "pwd", "ls -l", "whoami", "printenv | sort")
		//if err != nil {
		//	log.Errorf("Unable to create installation script %s for %s: %s", toolInstallScript, toolName, err)
		//	os.Exit(1)
		//}
		//shell.ExecuteScript(toolInstallScript)
	},
}
