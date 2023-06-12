package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	myos "github.com/nicholasdille/docker-setup/pkg/os"
)

var cronUpdateScript = `#!/bin/bash
set -o errexit

docker-setup update
docker-setup install --installed
`
var cronUpgradeScript = `#!/bin/bash
set -o errexit

outputPath="$(which docker-setup)"
curl https://github.com/nicholasdille/docker-setup/releases/latest/download/docker-setup \
	--location \
	--fail \
	--output "${outputPath}"
chmod +x "${outputPath}"
`

var create bool
var remove bool

func initCronCmd() {
	rootCmd.AddCommand(cronCmd)

	cronCmd.Flags().BoolVar(&create, "create", false, "Create cron jobs")
	cronCmd.Flags().BoolVar(&remove, "remove", false, "Remove cron jobs")
	cronCmd.MarkFlagsMutuallyExclusive("create", "remove")
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

func createCron() error {
	osVendor, err := myos.GetOsVendor(prefix)
	if err != nil {
		return fmt.Errorf("cannot determine OS: %w", err)
	}

	var cronWeeklyPath string
	var cronDailyPath string
	switch osVendor {
	case "ubuntu":
		cronWeeklyPath = "/etc/cron.weekly"
		cronDailyPath = "/etc/cron.daily"
	case "alpine":
		cronWeeklyPath = "/etc/periodic/weekly"
		cronDailyPath = "/etc/periodic/daily"
	default:
		return fmt.Errorf("unsupported OS: %s", osVendor)
	}

	// Write cronUpdateScript to /etc/cron.daily/docker-setup-update
	updateScript := []byte(cronUpdateScript)
	err = os.WriteFile(fmt.Sprintf("%s/docker-setup-update", cronWeeklyPath), updateScript, 0755)
	if err != nil {
		return fmt.Errorf("cannot write cron update script: %w", err)
	}

	// Write cronUpgradeScript to /etc/cron.weekly/docker-setup-upgrade
	upgradeScript := []byte(cronUpgradeScript)
	err = os.WriteFile(fmt.Sprintf("%s/docker-setup-upgrade", cronDailyPath), upgradeScript, 0755)
	if err != nil {
		return fmt.Errorf("cannot write cron upgrade script: %w", err)
	}

	return nil
}

func removeCron() error {
	// Check if exists /etc/cron.daily/docker-setup-update
	if fileExists(prefix + "/etc/cron.weekly/docker-setup-update") {
		// Remove /etc/cron.daily/docker-setup-update
		err := os.Remove(prefix + "/etc/cron.weekly/docker-setup-update")
		if err != nil {
			return fmt.Errorf("cannot remove cron update script: %w", err)
		}
	}

	// Check if exists /etc/cron.weekly/docker-setup-upgrade
	if fileExists(prefix + "/etc/cron.daily/docker-setup-upgrade") {
		// Remove /etc/cron.weekly/docker-setup-upgrade
		err := os.Remove(prefix + "/etc/cron.daily/docker-setup-upgrade")
		if err != nil {
			return fmt.Errorf("cannot remove cron upgrade script: %w", err)
		}
	}

	return nil
}
