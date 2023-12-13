package main

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"time"

	"github.com/pterm/pterm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/uniget-org/cli/pkg/logging"
	"github.com/uniget-org/cli/pkg/tool"
	"golang.org/x/sys/unix"
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

var altArch string = runtime.GOARCH
var arch string

var prefix = ""
var target = "usr/local"
var cacheRoot = "var/cache"
var cacheDirectory = cacheRoot + "/" + projectName
var libRoot = "var/lib"
var libDirectory = libRoot + "/" + projectName
var user = false
var metadataFileName = "metadata.json"
var metadataFile = cacheDirectory + "/" + metadataFileName
var registry = "ghcr.io"
var projectRepository = "uniget-org/cli"
var imageRepository = "uniget-org/tools"
var toolSeparator = "/"
var registryImagePrefix = registry + "/" + imageRepository + toolSeparator
var tools tool.Tools
var noInteractive bool

var (
	rootCmd = &cobra.Command{
		Use:          projectName,
		Version:      version,
		Short:        header + "The universal installer and updater to (container) tools",
		SilenceUsage: true,
	}
)

func directoryExists(directory string) bool {
	logging.Debug.Printfln("Checking if directory %s exists", directory)
	_, err := os.Stat(directory)
	return err == nil
}

func fileExists(file string) bool {
	logging.Debug.Printfln("Checking if file %s exists", file)
	_, err := os.Stat(file)
	return err == nil
}

func directoryIsWritable(directory string) bool {
	logging.Debug.Printfln("Checking if directory %s is writable", directory)
	return unix.Access(directory, unix.W_OK) == nil
}

func assertWritableDirectory(directory string) {
	if !directoryExists(directory) {
		assertDirectory(directory)
	}
	if !directoryIsWritable(directory) {
		logging.Error.Printfln("Directory %s is not writable", directory)
		os.Exit(1)
	}
}

func assertWritableTarget() {
	assertWritableDirectory(prefix + "/" + target)
}

func assertDirectory(directory string) {
	logging.Debug.Printfln("Creating directory %s", directory)
	err := os.MkdirAll(directory, 0755) // #nosec G301 -- Directories will contain public information
	if err != nil {
		logging.Error.Printfln("Error creating directory %s: %s", directory, err)
		os.Exit(1)
	}
}

func assertLibDirectory() {
	if !directoryExists(prefix + "/" + libRoot) {
		assertDirectory(prefix + "/" + libRoot)
	}
	assertWritableDirectory(prefix + "/" + libRoot)
	assertDirectory(prefix + "/" + libDirectory)
}

func assertCacheDirectory() {
	if !directoryExists(prefix + "/" + cacheRoot) {
		assertDirectory(prefix + "/" + cacheRoot)
	}
	assertWritableDirectory(prefix + "/" + cacheRoot)
	assertDirectory(prefix + "/" + cacheDirectory)
}

func assertMetadataFileExists() {
	_, err := os.Stat(prefix + "/" + metadataFile)
	if err != nil {
		logging.Error.Printfln("Metadata file %s does not exist: %s", prefix+"/"+metadataFile, err)
		os.Exit(1)
	}
}

func assertMetadataIsLoaded() {
	if len(tools.Tools) == 0 {
		logging.Error.Printfln("Metadata is not loaded")
		os.Exit(1)
	}
}

func init() {
	if altArch == "amd64" {
		arch = "x86_64"

	} else if altArch == "arm64" {
		arch = "aarch64"

	} else {
		logging.Error.Printfln("Unsupported architecture: %s", arch)
		os.Exit(1)
	}

	initCronCmd()
	initDescribeCmd()
	initGenerateCmd()
	initHealthcheckCmd()
	initInspectCmd()
	initInstallCmd()
	initListCmd()
	initMessageCmd()
	initPostinstallCmd()
	initSearchCmd()
	initSelfUpgradeCmd()
	initTagsCmd()
	initUninstallCmd()
	initUpdateCmd()
	initUpgradeCmd()
	initVersionCmd()
}

func main() {
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		logging.Error.Writer = os.Stderr
		pterm.Warning.Writer = os.Stderr

		if trace {
			pterm.EnableDebugMessages()
			log.SetLevel(log.TraceLevel)

		} else if debug {
			pterm.EnableDebugMessages()
			log.SetLevel(log.DebugLevel)

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
			target = os.Getenv("HOME") + "/.local"
			cacheRoot = os.Getenv("HOME") + "/.cache"
			cacheDirectory = cacheRoot + "/" + projectName
			libRoot = os.Getenv("HOME") + "/.local/state"
			libDirectory = libRoot + "/" + projectName
			metadataFile = cacheDirectory + "/" + metadataFileName

		} else {
			cacheDirectory = cacheRoot + "/" + cacheDirectory
			libDirectory = libRoot + "/" + libDirectory
			metadataFile = cacheDirectory + "/" + metadataFileName
		}

		if debug {
			logging.Debug.Printfln("target: %s", target)
			logging.Debug.Printfln("cacheRoot: %s", cacheRoot)
			logging.Debug.Printfln("cacheDirectory: %s", cacheDirectory)
			logging.Debug.Printfln("libRoot: %s", libRoot)
			logging.Debug.Printfln("libDirectory: %s", libDirectory)
			logging.Debug.Printfln("metadataFile: %s", metadataFile)
		}

		if !fileExists(prefix + "/" + metadataFile) {
			logging.Debug.Printfln("Metadata file does not exist. Downloading...")
			err := downloadMetadata()
			if err != nil {
				return fmt.Errorf("error downloading metadata: %s", err)
			}
		} else {
			logging.Debug.Printfln("Metadata file exists")
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
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", log.WarnLevel.String(), "Log level (trace, debug, info, warning, error)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Set log level to debug")
	rootCmd.PersistentFlags().BoolVar(&trace, "trace", false, "Set log level to trace")
	rootCmd.PersistentFlags().StringVarP(&prefix, "prefix", "p", "", "Base directory for the installation (useful when preparing a chroot environment)")
	rootCmd.PersistentFlags().StringVarP(&target, "target", "t", "usr/local", "Target directory for installation relative to PREFIX")
	rootCmd.PersistentFlags().StringVarP(&cacheRoot, "cache-root", "C", "var/cache", "Cache root directory relative to PREFIX")
	rootCmd.PersistentFlags().StringVarP(&cacheDirectory, "cache-directory", "c", projectName, "Cache directory relative to CACHE-ROOT")
	rootCmd.PersistentFlags().StringVarP(&libRoot, "lib-root", "L", "var/lib", "Library root directory relative to PREFIX")
	rootCmd.PersistentFlags().StringVarP(&libDirectory, "lib-directory", "l", projectName, "Library directory relative to LIB-ROOT")
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
