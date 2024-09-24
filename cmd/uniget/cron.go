package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var create bool
var createUpgradeHour string
var createSelfUpgradeHour string
var createSelfUpgradeDay string
var remove bool

func initCronCmd() {
	rootCmd.AddCommand(cronCmd)

	cronCmd.Flags().BoolVar(&create, "create", false, "Create cron jobs")
	cronCmd.Flags().StringVar(&createUpgradeHour, "upgrade-hour", "1", "Hour to run cron jobs for tool upgrade")
	cronCmd.Flags().StringVar(&createSelfUpgradeHour, "self-upgrade-hour", "0", "Hour to run cron jobs for self-upgrade")
	cronCmd.Flags().StringVar(&createSelfUpgradeDay, "self-upgrade-day", "0", "Day to run cron jobs for self-upgrade on")
	cronCmd.Flags().BoolVar(&remove, "remove", false, "Remove cron jobs")
	cronCmd.MarkFlagsMutuallyExclusive("create", "remove")
	cronCmd.MarkFlagsMutuallyExclusive("remove", "upgrade-hour")
	cronCmd.MarkFlagsMutuallyExclusive("remove", "self-upgrade-hour")
}

var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "Create cron jobs",
	Long:  header + "\nCreate cron jobs for updating",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if create {
			return createCron()
		}
		if remove {
			return removeCron()
		}

		return fmt.Errorf("either --create or --remove must be specified")
	},
}

func getUserCrontab() ([]string, error) {
	cmd := exec.Command("crontab", "-l")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("cannot get user crontab: %w", err)
	}

	lines := []string{}
	if len(output) > 0 {
		lines = strings.Split(string(output), "\n")
	}

	return lines, nil
}

func removeUserCronTab(lines []string) []string {
	newLines := []string{}

	for i := len(lines) - 1; i >= 0; i-- {
		if !strings.Contains(lines[i], "uniget") {
			newLines = append(newLines, lines[i])
		}
	}

	return newLines
}

func setUserCrontab(lines []string) error {
	input := strings.Join(lines, "\n") + "\n"
	if len(lines) == 0 {
		input = ""
	}
	cmd := exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(input)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cannot set user crontab: %s", output)
	}

	return nil
}

func createCron() error {
	lines, err := getUserCrontab()
	if err != nil {
		return fmt.Errorf("cannot get user crontab: %w", err)
	}
	lines = removeUserCronTab(lines)
	lines = append(lines, fmt.Sprintf("30 %s * * * uniget --user=%t update && uniget --user=%t install --installed", createUpgradeHour, viper.GetBool("user"), viper.GetBool("user")))
	lines = append(lines, fmt.Sprintf("0 %s * * %s uniget --user=%t self-upgrade", createSelfUpgradeHour, createSelfUpgradeDay, viper.GetBool("user")))

	err = setUserCrontab(lines)
	if err != nil {
		return fmt.Errorf("cannot set user crontab: %w", err)
	}

	return nil
}

func removeCron() error {
	lines, err := getUserCrontab()
	if err != nil {
		return fmt.Errorf("cannot get user crontab: %w", err)
	}
	lines = removeUserCronTab(lines)
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	err = setUserCrontab(lines)
	if err != nil {
		return fmt.Errorf("cannot set user crontab: %w", err)
	}

	return nil
}
