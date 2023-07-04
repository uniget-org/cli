package main

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/pterm/pterm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
var debug bool
var trace bool

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
	initMessageCmd()
	initPostinstallCmd()
	initSearchCmd()
	initTagsCmd()
	initUninstallCmd()
	initUpdateCmd()
	initVersionCmd()
}

func main() {
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if debug {
			pterm.EnableDebugMessages()

		} else if trace {
			pterm.EnableDebugMessages()
			log.SetLevel(log.TraceLevel)

		} else {
			log.SetLevel(log.WarnLevel)
		}

		if len(prefix) > 0 {
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
				log.Debugf("Converted prefix to absolute path %s\n", prefix)
			}
		}

		if user {
			pterm.Debug.Println("Installing in user context")
			target = os.Getenv("HOME") + "/.local/bin"
			cacheRoot = os.Getenv("HOME") + "/.cache"
			cacheDirectory = cacheRoot + "/docker-setup"
			libRoot = os.Getenv("HOME") + "/.local/state"
			libDirectory = libRoot + "/docker-setup"

		} else {
			cacheDirectory = cacheRoot + "/" + cacheDirectory
			libDirectory = libRoot + "/" + libDirectory
		}

		if !fileExists(prefix + "/" + metadataFile) {
			err := downloadMetadata()
			if err != nil {
				return fmt.Errorf("error downloading metadata: %s", err)
			}
		}
		err := loadMetadata()
		if err != nil {
			return fmt.Errorf("error loading metadata: %s", err)
		}

		file, err := os.Stat(prefix + "/" + metadataFile)
		if err != nil {
			return fmt.Errorf("error stating metadata file: %s", err)
		}
		now := time.Now()
		modifiedtime := file.ModTime()
		if now.Sub(modifiedtime).Hours() > 24 {
			pterm.Warning.Println("Metadata file is older than 24 hours")
		}

		return nil
	}
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", log.WarnLevel.String(), "Log level (trace, debug, info, warning, error)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Set log level to debug")
	rootCmd.PersistentFlags().BoolVar(&trace, "trace", false, "Set log level to trace")
	rootCmd.PersistentFlags().StringVarP(&prefix, "prefix", "p", "", "Prefix for installation")
	rootCmd.PersistentFlags().StringVarP(&target, "target", "t", "usr/local", "Target directory for installation")
	rootCmd.PersistentFlags().StringVarP(&cacheRoot, "cache-root", "C", "var/cache", "Cache root directory relative to PREFIX")
	rootCmd.PersistentFlags().StringVar(&cacheDirectory, "cache-directory", "docker-setup", "Cache directory relative to CACHE-ROOT")
	rootCmd.PersistentFlags().StringVarP(&libRoot, "lib-root", "L", "var/lib", "Library root directory relative to PREFIX")
	rootCmd.PersistentFlags().StringVar(&libDirectory, "lib-directory", "docker-setup", "Library directory relative to LIB-ROOT")
	rootCmd.PersistentFlags().BoolVarP(&user, "user", "u", false, "Install in user context")
	rootCmd.PersistentFlags().StringVarP(&metadataFileName, "metadata-file", "f", "metadata.json", "Metadata file")
	rootCmd.PersistentFlags().BoolVar(&noInteractive, "no-interactive", false, "Disable interactive prompts")

	rootCmd.MarkFlagsMutuallyExclusive("prefix", "user")
	rootCmd.MarkFlagsMutuallyExclusive("target", "user")
	rootCmd.MarkFlagsMutuallyExclusive("cache-directory", "user")
	rootCmd.MarkFlagsMutuallyExclusive("lib-directory", "user")

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
