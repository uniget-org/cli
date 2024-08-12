package main

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	goversion "github.com/hashicorp/go-version"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
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
var configRoot = "etc"
var profileDDirectory = configRoot + "/profile.d"
var metadataFileName = "metadata.json"
var metadataFile = cacheDirectory + "/" + metadataFileName
var registry = "ghcr.io"
var projectRepository = "uniget-org/cli"
var imageRepository = "uniget-org/tools"
var toolSeparator = "/"
var registryImagePrefix = registry + "/" + imageRepository + toolSeparator
var tools tool.Tools

var pathRewriteRules = make([]tool.PathRewrite, 0)

var (
	rootCmd = &cobra.Command{
		Use:          projectName,
		Version:      version,
		Short:        header + "The universal installer and updater to (container) tools",
		SilenceUsage: true,
	}
)

var minimumCliVersionForSchemaVersion = map[string]string{
	"1": "0.1.0",
}

func checkClientVersionRequirement(tool *tool.Tool) {
	if version == "main" {
		logging.Warning.Printfln("You are running an unreleased version of uniget. Cannot check client version requirement for %s", tool.Name)
		return
	}

	var requiredCliVersion = "0.0.0"
	for schemaVersion, cliVersion := range minimumCliVersionForSchemaVersion {
		if tool.SchemaVersion > schemaVersion {
			requiredCliVersion = cliVersion
		}
	}

	logging.Debugf("Checking if client version %s is at least %s", version, requiredCliVersion)
	
    v1, err := goversion.NewVersion(requiredCliVersion)
    if err != nil {
        panic(err)
    }
    v2, err := goversion.NewVersion(version)
    if err != nil {
        panic(err)
    }

    if v1.GreaterThan(v2) {
		logging.Error.Printfln("The tool %s requires at least version %s but you have %s", tool.Name, requiredCliVersion, version)
		os.Exit(1)
	}
}

func directoryExists(directory string) bool {
	logging.Debugf("Checking if directory %s exists", directory)
	_, err := os.Stat(directory)
	return err == nil
}

func fileExists(file string) bool {
	logging.Debugf("Checking if file %s exists", file)
	_, err := os.Stat(file)
	return err == nil
}

func directoryIsWritable(directory string) bool {
	logging.Debugf("Checking if directory %s is writable", directory)
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
	logging.Debugf("Creating directory %s", directory)
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
	initDebugCmd()
	initDescribeCmd()
	initEnvCmd()
	initGenerateCmd()
	initHealthcheckCmd()
	initInspectCmd()
	initInstallCmd()
	initListCmd()
	initManCmd()
	initMessageCmd()
	initReleaseNotesCmd()
	initSearchCmd()
	initSelfUpgradeCmd()
	initShimCmd()
	initTagsCmd()
	initUninstallCmd()
	initUpdateCmd()
	initUpgradeCmd()
	initVersionCmd()
}

func addViperBindings(flags *flag.FlagSet, cobraLongName string, viperName string) {
	err := viper.BindPFlag(viperName, flags.Lookup(cobraLongName))
	if err != nil {
		fmt.Printf("unable to bind flag %s: %s", cobraLongName, err)
		os.Exit(1)
	}

	if viperName != cobraLongName {
		err = viper.BindEnv(viperName, strings.ToUpper(viper.GetEnvPrefix() + "_" + strings.ReplaceAll(cobraLongName, "-", "_")))
		if err != nil {
			fmt.Printf("unable to bind environment variable for flag %s: %s", cobraLongName, err)
			os.Exit(1)
		}
	}
}

func main() {
	var err error

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		logging.Error.Writer = os.Stderr
		logging.Warning.Writer = os.Stderr

		if viper.GetBool("trace") {
			pterm.EnableDebugMessages()
			pterm.DefaultLogger.Level = pterm.LogLevelTrace
			logging.Level = pterm.LogLevelTrace

		} else if viper.GetBool("debug") {
			pterm.EnableDebugMessages()
			pterm.DefaultLogger.Level = pterm.LogLevelDebug
			logging.Level = pterm.LogLevelDebug

		} else {
			pterm.DisableDebugMessages()
			pterm.DefaultLogger.Level = pterm.LogLevelInfo
			logging.Level = pterm.LogLevelInfo
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
				logging.Debugf("Converted prefix to absolute path %s", viper.GetString("prefix"))
			}
		}

		if viper.GetBool("user") {
			viper.Set("prefix", os.Getenv("HOME"))
			viper.Set("target", ".local")

			cacheRoot = ".cache"
			if os.Getenv("XDG_CACHE_HOME") != "" {
				if strings.HasPrefix(os.Getenv("XDG_CACHE_HOME"), os.Getenv("HOME")) {
					cacheRoot = strings.TrimPrefix(os.Getenv("XDG_CACHE_HOME"), os.Getenv("HOME")+"/")
				}
			}
			cacheDirectory = cacheRoot + "/" + projectName

			libRoot = ".local/state"
			if os.Getenv("XDG_STATE_HOME") != "" {
				if strings.HasPrefix(os.Getenv("XDG_STATE_HOME"), os.Getenv("HOME")) {
					libRoot = strings.TrimPrefix(os.Getenv("XDG_STATE_HOME"), os.Getenv("HOME")+"/")
				}
			}
			libDirectory = libRoot + "/" + projectName

			configRoot = ".config"
			if os.Getenv("XDG_CONFIG_HOME") != "" {
				if strings.HasPrefix(os.Getenv("XDG_CONFIG_HOME"), os.Getenv("HOME")) {
					configRoot = strings.TrimPrefix(os.Getenv("XDG_CONFIG_HOME"), os.Getenv("HOME")+"/")
				}
			}
			profileDDirectory = configRoot + "/profile.d"

			metadataFile = cacheDirectory + "/" + metadataFileName
		}

		if strings.HasPrefix(viper.GetString("target"), "/") {
			viper.Set("target", strings.TrimLeft(viper.GetString("target"), "/"))
		}

		if viper.GetBool("debug") {
			logging.Debugf("user: %t", viper.GetBool("prefix"))
			logging.Debugf("prefix: %s", viper.GetString("prefix"))
			logging.Debugf("target: %s", viper.GetString("target"))
			logging.Debugf("cacheRoot: %s", cacheRoot)
			logging.Debugf("cacheDirectory: %s", cacheDirectory)
			logging.Debugf("libRoot: %s", libRoot)
			logging.Debugf("libDirectory: %s", libDirectory)
			logging.Debugf("metadataFile: %s", metadataFile)
			logging.Debugf("use-path-rewrite: %t", viper.GetBool("usepathrewrite"))
		}

		pathRewriteRules = []tool.PathRewrite{
			{
				Source: "usr/local/",
				Target: "",
				Operation: "REPLACE",
			},
			{
				Source: "var/lib/uniget/",
				Target: libDirectory + "/",
				Operation: "REPLACE",
			},
			{
				Source: "var/cache/uniget/",
				Target: cacheDirectory + "/",
				Operation: "REPLACE",
			},
		}
		if ! viper.GetBool("user") {
			logging.Debugf("Adding path rewrite rules for system installation")

			if viper.GetBool("integratesystemd") || viper.GetBool("integrateall") {
				pathRewriteRules = append(pathRewriteRules, tool.PathRewrite{
					Source: "etc/systemd/",
					Target: "/etc/systemd/",
					Operation: "REPLACE",
				})
			}

			if viper.GetBool("integrateprofiled") || viper.GetBool("integrateall") {
				pathRewriteRules = append(pathRewriteRules, tool.PathRewrite{
					Source: "etc/profile.d/",
					Target: "/etc/profile.d/",
					Operation: "REPLACE",
				})
			}

		} else {
			logging.Debugf("Adding path rewrite rules for user installation")

			if viper.GetBool("integratedockercliplugins") || viper.GetBool("integrateall") {
				pathRewriteRules = append(pathRewriteRules, tool.PathRewrite{
					Source: "libexec/docker/cli-plugins/",
					Target: "./.docker/cli-plugins/",
					Operation: "REPLACE",
				})
			}

			if viper.GetBool("integratesystemd") || viper.GetBool("integrateall") {
				pathRewriteRules = append(pathRewriteRules, tool.PathRewrite{
					Source: "etc/systemd/user/",
					Target: "./.config/systemd/user/",
					Operation: "REPLACE",
				})
			}

			if viper.GetBool("integrateprofiled") || viper.GetBool("integrateall") {
				pathRewriteRules = append(pathRewriteRules, tool.PathRewrite{
					Source: "etc/profile.d/",
					Target: "./.config/profile.d/",
					Operation: "REPLACE",
				})
			}

			if viper.GetBool("integrateetc") || viper.GetBool("integrateall") {
				pathRewriteRules = append(pathRewriteRules, tool.PathRewrite{
					Source: "etc/",
					Target: "./.config/",
					Operation: "REPLACE",
				})
			}
		}
		pathRewriteRules = append(pathRewriteRules, tool.PathRewrite{
			Source: "",
			Target: viper.GetString("target") + "/",
			Operation: "PREPEND",
		})
		if viper.GetBool("debug") {
			logging.Debug("Path rewrite rules:")
			for _, rule := range pathRewriteRules {
				logging.Debugf("  %s -> %s (%s)", rule.Source, rule.Target, rule.Operation)
			}
		}

		if !fileExists(viper.GetString("prefix") + "/" + metadataFile) {
			logging.Debugf("Metadata file does not exist. Downloading...")
			err := downloadMetadata()
			if err != nil {
				return fmt.Errorf("error downloading metadata: %s", err)
			}
		} else {
			logging.Debugf("Metadata file exists")
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
			logging.Warning.Println("Metadata file is older than 24 hours")
		}

		return nil
	}

	viper.SetDefault("loglevel", pterm.LogLevelInfo.String())
	viper.SetDefault("debug", false)
	viper.SetDefault("trace", false)
	viper.SetDefault("prefix", "")
	viper.SetDefault("target", "usr/local")
	viper.SetDefault("user", false)
	viper.SetDefault("autoupdate", false)
	viper.SetDefault("integratesystemd", false)
	viper.SetDefault("integrateprofiled", false)
	viper.SetDefault("integrateetc", false)
	viper.SetDefault("integratedockercliplugins", false)
	viper.SetDefault("integrateall", false)
	viper.SetDefault("usepathrewrite", false)

	pf := rootCmd.PersistentFlags()

	pf.String("log-level", viper.GetString("loglevel"), "Log level (trace, debug, info, warning, error)")
	pf.BoolP("debug", "d", viper.GetBool("debug"), "Set log level to debug")
	pf.Bool("trace", viper.GetBool("trace"), "Set log level to trace")
	pf.StringP("prefix", "p", viper.GetString("prefix"), "Base directory for the installation (useful when preparing a chroot environment)")
	pf.StringP("target", "t", viper.GetString("target"), "Target directory for installation relative to PREFIX")
	pf.BoolP("user", "u", viper.GetBool("user"), "Install in user context")
	pf.Bool("auto-update", viper.GetBool("autoupdate"), "Automatically update metadata")
	pf.Bool("integrate-systemd", viper.GetBool("integratesystemd"), "Integrate systemd unit files")
	pf.Bool("integrate-profiled", viper.GetBool("integrateprofiled"), "Integrate profile.d scripts")
	pf.Bool("integrate-etc", viper.GetBool("integrateetc"), "Integrate configuration files from /etc")
	pf.Bool("integrate-docker-cli-plugins", viper.GetBool("integratedockercliplugins"), "Integrate Docker CLI plugins")
	pf.Bool("integrate-all", viper.GetBool("integrateall"), "Integrate all available integrations")
	pf.Bool("use-path-rewrite", viper.GetBool("usepathrewrite"), "(Experimental) Enable path rewrite rules for installation")

	rootCmd.MarkFlagsMutuallyExclusive("prefix", "user")
	rootCmd.MarkFlagsMutuallyExclusive("target", "user")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("uniget")

	addViperBindings(pf, "log-level", "loglevel")
	addViperBindings(pf, "debug", "debug")
	addViperBindings(pf, "trace", "trace")
	addViperBindings(pf, "prefix", "prefix")
	addViperBindings(pf, "target", "target")
	addViperBindings(pf, "user", "user")
	addViperBindings(pf, "auto-update", "autoupdate")
	addViperBindings(pf, "integrate-systemd", "integratesystemd")
	addViperBindings(pf, "integrate-profiled", "integrateprofiled")
	addViperBindings(pf, "integrate-etc", "integrateetc")
	addViperBindings(pf, "integrate-docker-cli-plugins", "integratedockercliplugins")
	addViperBindings(pf, "integrate-all", "integrateall")
	addViperBindings(pf, "use-path-rewrite", "usepathrewrite")

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
