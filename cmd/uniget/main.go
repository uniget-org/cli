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
	"github.com/spf13/viper"
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

var altArch string = runtime.GOARCH
var arch string

var cacheRoot = "var/cache"
var cacheDirectory = cacheRoot + "/" + projectName
var libRoot = "var/lib"
var libDirectory = libRoot + "/" + projectName
var metadataFileName = "metadata.json"
var metadataFile = cacheDirectory + "/" + metadataFileName
var registry = "ghcr.io"
var projectRepository = "uniget-org/cli"
var imageRepository = "uniget-org/tools"
var toolSeparator = "/"
var registryImagePrefix = registry + "/" + imageRepository + toolSeparator
var tools tool.Tools

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
	assertWritableDirectory(viper.GetString("prefix") + "/" + viper.GetString("target"))
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
	if !directoryExists(viper.GetString("prefix") + "/" + libRoot) {
		assertDirectory(viper.GetString("prefix") + "/" + libRoot)
	}
	assertWritableDirectory(viper.GetString("prefix") + "/" + libRoot)
	assertDirectory(viper.GetString("prefix") + "/" + libDirectory)
}

func assertCacheDirectory() {
	if !directoryExists(viper.GetString("prefix") + "/" + cacheRoot) {
		assertDirectory(viper.GetString("prefix") + "/" + cacheRoot)
	}
	assertWritableDirectory(viper.GetString("prefix") + "/" + cacheRoot)
	assertDirectory(viper.GetString("prefix") + "/" + cacheDirectory)
}

func assertMetadataFileExists() {
	_, err := os.Stat(viper.GetString("prefix") + "/" + metadataFile)
	if err != nil {
		logging.Error.Printfln("Metadata file %s does not exist: %s", viper.GetString("prefix")+"/"+metadataFile, err)
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

		if viper.GetBool("trace") {
			pterm.EnableDebugMessages()
			log.SetLevel(log.TraceLevel)

		} else if viper.GetBool("debug") {
			pterm.EnableDebugMessages()
			log.SetLevel(log.DebugLevel)

		} else {
			log.SetLevel(log.WarnLevel)
		}

		if len(viper.GetString("prefix")) > 0 {
			re, err := regexp.Compile(`^\/`)
			if err != nil {
				return fmt.Errorf("cannot compile regexp: %w", err)
			}
			if !re.MatchString(viper.GetString("prefix")) {
				wd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("cannot determine working directory: %w", err)
				}
				viper.Set("prefix", wd+"/"+viper.GetString("prefix"))
				log.Debugf("Converted prefix to absolute path %s\n", viper.GetString("prefix"))
			}
		}

		if viper.GetBool("user") {
			viper.Set("prefix", os.Getenv("HOME"))
			viper.Set("target", ".local")
			cacheRoot = ".cache"
			cacheDirectory = cacheRoot + "/" + projectName
			libRoot = ".local/state"
			libDirectory = libRoot + "/" + projectName
			metadataFile = cacheDirectory + "/" + metadataFileName
		}

		if viper.GetBool("debug") {
			logging.Debug.Printfln("target: %s", viper.GetString("target"))
			logging.Debug.Printfln("cacheRoot: %s", cacheRoot)
			logging.Debug.Printfln("cacheDirectory: %s", cacheDirectory)
			logging.Debug.Printfln("libRoot: %s", libRoot)
			logging.Debug.Printfln("libDirectory: %s", libDirectory)
			logging.Debug.Printfln("metadataFile: %s", metadataFile)
		}

		if !fileExists(viper.GetString("prefix") + "/" + metadataFile) {
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

		file, err := os.Stat(viper.GetString("prefix") + "/" + metadataFile)
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

	viper.SetDefault("log-level", log.WarnLevel.String())
	viper.SetDefault("debug", false)
	viper.SetDefault("trace", false)
	viper.SetDefault("prefix", "")
	viper.SetDefault("target", "usr/local")
	viper.SetDefault("user", false)
	viper.SetDefault("no-interactive", false)

	pf := rootCmd.PersistentFlags()

	pf.String("log-level", viper.GetString("log-level"), "Log level (trace, debug, info, warning, error)")
	pf.BoolP("debug", "d", viper.GetBool("debug"), "Set log level to debug")
	pf.Bool("trace", viper.GetBool("trace"), "Set log level to trace")
	pf.StringP("prefix", "p", viper.GetString("prefix"), "Base directory for the installation (useful when preparing a chroot environment)")
	pf.StringP("target", "t", viper.GetString("target"), "Target directory for installation relative to PREFIX")
	pf.BoolP("user", "u", viper.GetBool("user"), "Install in user context")
	pf.Bool("no-interactive", viper.GetBool("no-interactive"), "Disable interactive prompts")

	rootCmd.MarkFlagsMutuallyExclusive("prefix", "user")
	rootCmd.MarkFlagsMutuallyExclusive("target", "user")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("uniget")
	viper.BindPFlag("log-level", pf.Lookup("log-level"))
	viper.BindPFlag("debug", pf.Lookup("debug"))
	viper.BindPFlag("trace", pf.Lookup("trace"))
	viper.BindPFlag("prefix", pf.Lookup("prefix"))
	viper.BindPFlag("target", pf.Lookup("target"))
	viper.BindPFlag("user", pf.Lookup("user"))
	viper.BindPFlag("no-interactive", pf.Lookup("no-interactive"))

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
