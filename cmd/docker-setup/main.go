package main

import (
	"fmt"
	"os"
	"regexp"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	git "github.com/go-git/go-git/v5"
)

var version string = "main"
var header string = `
     _             _
    | |           | |                                  _
  __| | ___   ____| |  _ _____  ____ _____ ___ _____ _| |_ _   _ ____
 / _  |/ _ \ / ___) |_/ ) ___ |/ ___|_____)___) ___ (_   _) | | |  _ \
( (_| | |_| ( (___|  _ (| ____| |        |___ | ____| | |_| |_| | |_| |
 \____|\___/ \____)_| \_)_____)_|        (___/|_____)  \__)____/|  __/
                                                                |_|
`
var logLevel string

var (
	rootCmd = &cobra.Command{
		Use:          "docker-setup",
		Version:      version,
		Short:        header + "The container tools installer and updater",
		SilenceUsage: true,
	}
)

func init() {
	initDockerSetup()

	initCronCmd()
	initDescribeCmd()
	initGenerateCmd()
	initInspectCmd()
	initInstallCmd()
	initListCmd()
	initSearchCmd()
	initTagsCmd()
	initUninstallCmd()
	initUpdateCmd()
	initVersionCmd()

	if fileExists(".git/config") {
		repo, err := git.PlainOpen(".")
		if err != nil {
			log.Fatal(err)
		}
		config, err := repo.Config()
		if err != nil {
			log.Fatal(err)
		}
		origin := config.Remotes["origin"]
		if origin.URLs[0] == "https://github.com/nicholasdille/docker-setup" {
			initDevCmd()
		}
	}
}

func main() {
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		log.SetOutput(os.Stdout)
		level, err := log.ParseLevel(logLevel)
		if err != nil {
			return err
		}
		log.SetLevel(level)
		log.Debugf("Log level is now %s\n", logLevel)

		re, err := regexp.Compile(`^\/`)
		if err != nil {
			return fmt.Errorf("cannot compile regexp: %w", err)
		}
		if !re.MatchString(prefix) {
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("cannot determine working directory: %w", err)
			}
			prefix = wd + "/" + prefix
			log.Debugf("Convered prefix to absolute path %s\n", prefix)
		}

		return nil
	}
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", log.WarnLevel.String(), "Log level (trace, debug, info, warning, error)")
	// TODO: Add flags --trace and --debug (make mutually exclusive)
	rootCmd.PersistentFlags().StringVarP(&prefix, "prefix", "p", "/", "Prefix for installation")
	rootCmd.PersistentFlags().StringVarP(&target, "target", "t", "usr/local", "Target directory for installation")
	rootCmd.PersistentFlags().StringVarP(&cacheDirectory, "cache-directory", "C", "var/cache/docker-setup", "Cache directory relative to PREFIX")
	rootCmd.PersistentFlags().StringVarP(&libDirectory, "lib-directory", "L", "var/lib/docker-setup", "Library directory relative to PREFIX")
	rootCmd.PersistentFlags().StringVarP(&metadataFileName, "metadata-file", "f", "metadata.json", "Metadata file")

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
