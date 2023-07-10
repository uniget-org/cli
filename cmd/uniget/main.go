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

var projectName = "uniget"
var version string = "main"
var header string = "" +
	"             _            _\n" +
	" _   _ _ __ (_) __ _  ___| |_\n" +
	"| | | | '_ \\| |/ _` |/ _ \\ __|\n" +
	"| |_| | | | | | (_| |  __/ |_\n" +
	" \\__,_|_| |_|_|\\__, |\\___|\\__|\n" +
	"               |___/\n"
var logLevel string
var debug bool
var trace bool

var (
	rootCmd = &cobra.Command{
		Use:          projectName,
		Version:      version,
		Short:        header + "The universal installer and updater to (container) tools",
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
	initUpgradeCmd()
	initVersionCmd()
}

func main() {
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		pterm.Error.Writer = os.Stderr
		pterm.Warning.Writer = os.Stderr

		if debug {
			pterm.EnableDebugMessages()
			log.SetLevel(log.DebugLevel)

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
			cacheDirectory = cacheRoot + "/" + projectName
			libRoot = os.Getenv("HOME") + "/.local/state"
			libDirectory = libRoot + "/" + projectName
			metadataFile = cacheDirectory + "/" + metadataFileName
			pterm.Error.Println("User context is not yet supported. Please check #6270.")

		} else {
			cacheDirectory = cacheRoot + "/" + cacheDirectory
			libDirectory = libRoot + "/" + libDirectory
		}

		if debug {
			pterm.Debug.Printfln("target: %s", target)
			pterm.Debug.Printfln("cacheRoot: %s", cacheRoot)
			pterm.Debug.Printfln("cacheDirectory: %s", cacheDirectory)
			pterm.Debug.Printfln("libRoot: %s", libRoot)
			pterm.Debug.Printfln("libDirectory: %s", libDirectory)
			pterm.Debug.Printfln("metadataFile: %s", metadataFile)
		}

		if !fileExists(prefix + "/" + metadataFile) {
			pterm.Debug.Printfln("Metadata file does not exist. Downloading...")
			err := downloadMetadata()
			if err != nil {
				return fmt.Errorf("error downloading metadata: %s", err)
			}
		} else {
			pterm.Debug.Printfln("Metadata file exists")
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
	rootCmd.PersistentFlags().StringVarP(&cacheDirectory, "cache-directory", "c", projectName, "Cache directory relative to CACHE-ROOT")
	rootCmd.PersistentFlags().StringVarP(&libRoot, "lib-root", "L", "var/lib", "Library root directory relative to PREFIX")
	rootCmd.PersistentFlags().StringVar(&libDirectory, "lib-directory", projectName, "Library directory relative to LIB-ROOT")
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
