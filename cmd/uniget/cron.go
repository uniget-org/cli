package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	createUpgradeCron     = "30 0 * * *"
	createSelfUpgradeCron = "0 0 * * *"
)

func initCronCmd() {
	rootCmd.AddCommand(cronCmd)

	cronCmd.AddCommand(cronCreateCmd)
	cronCmd.AddCommand(cronRemoveCmd)

	cronCreateCmd.Flags().StringVar(&createUpgradeCron, "upgrade-cron", createUpgradeCron, "Cron schedule to run cron jobs for tool upgrade")
	cronCreateCmd.Flags().StringVar(&createSelfUpgradeCron, "self-upgrade-cron", createSelfUpgradeCron, "Cron schedule to run cron jobs for self-upgrade")
}

var cronCmd = &cobra.Command{
	Use: "cron",
	Aliases: []string{
		"schedule",
		"s",
	},
	Short: "Manage cron jobs",
	Long:  header + "\nManage cron jobs for updating",
	Args:  cobra.NoArgs,
}

var cronCreateCmd = &cobra.Command{
	Use: "create",
	Aliases: []string{
		"c",
	},
	Short: "Create cron job",
	Long:  header + "\nCreate cron job",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return createCron()
	},
}

var cronRemoveCmd = &cobra.Command{
	Use: "remove",
	Aliases: []string{
		"r",
	},
	Short: "Remove cron job",
	Long:  header + "\nRemove cron job",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return removeCron()
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
	lines = append(lines, fmt.Sprintf("%s uniget --user=%t upgrade --auto-update", createUpgradeCron, viper.GetBool("user")))
	lines = append(lines, fmt.Sprintf("%s uniget --user=%t self-upgrade", createSelfUpgradeCron, viper.GetBool("user")))

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
