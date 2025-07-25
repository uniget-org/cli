package main

import (
	_ "embed"
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
	"github.com/uniget-org/cli/pkg/cache"
	"github.com/uniget-org/cli/pkg/containers"
	"github.com/uniget-org/cli/pkg/logging"
	"github.com/uniget-org/cli/pkg/tool"
	"golang.org/x/sys/unix"
)

var (
	projectName        = "uniget"
	version     string = "main"

	//go:embed logo.txt
	header string
	slogan string = "The universal installer and updater for (container) tools" + "\n" +
		"                                       https://uniget.dev"

	altArch string = runtime.GOARCH
	arch    string

	cacheRoot              = "var/cache"
	cacheDirectory         = cacheRoot + "/" + projectName
	libRoot                = "var/lib"
	libDirectory           = libRoot + "/" + projectName
	configRoot             = "etc"
	profileDDirectory      = configRoot + "/profile.d"
	metadataImageTag       = "main"
	metadataFileName       = "metadata.json"
	metadataFile           = cacheDirectory + "/" + metadataFileName
	fileCacheDirectoryName = "downloads"
	registry               = "ghcr.io"
	organization           = "uniget-org"
	imageRepository        = organization + "/tools"
	toolSeparator          = "/"
	registryImagePrefix    = registry + "/" + imageRepository + toolSeparator
	tools                  = tool.Tools{
		Tools: make([]tool.Tool, 0),
	}
	pathRewriteRules = make([]tool.PathRewrite, 0)
	rootCmd          = &cobra.Command{
		Use:          projectName,
		Version:      version,
		Short:        header + slogan,
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logging.OutputWriter = cmd.OutOrStdout()
			logging.ErrorWriter = cmd.ErrOrStderr()

			if viper.GetBool("trace") {
				pterm.EnableDebugMessages()
				logging.Level = pterm.LogLevelTrace

			} else if viper.GetBool("debug") {
				pterm.EnableDebugMessages()
				logging.Level = pterm.LogLevelDebug

			} else {
				pterm.DisableDebugMessages()
				logging.Level = pterm.LogLevelInfo
			}

			logging.Init()

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

			if !viper.GetBool("user") {
				cacheDirectory = cacheRoot + "/" + projectName
				libDirectory = libRoot + "/" + projectName
				profileDDirectory = configRoot + "/profile.d"
				metadataFile = cacheDirectory + "/" + metadataFileName
				viper.Set("cachedirectory", cacheDirectory+"/"+fileCacheDirectoryName)

			} else {
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
				viper.Set("cachedirectory", cacheDirectory+"/"+fileCacheDirectoryName)
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
				logging.Debugf("registry: %s", viper.GetString("registry"))
				logging.Debugf("repository: %s", viper.GetString("repository"))
				logging.Debugf("tool-separator: %s", viper.GetString("toolseparator"))
				logging.Debugf("cache: %s", viper.GetString("cache"))
				logging.Debugf("cache-directory: %s", viper.GetString("cachedirectory"))
				logging.Debugf("cache-retention: %d", viper.GetInt("cacheretention"))
			}

			pathRewriteRules = []tool.PathRewrite{
				{
					Source:    "usr/local/",
					Target:    "",
					Operation: "REPLACE",
				},
				{
					Source:    "var/lib/uniget/",
					Target:    libDirectory + "/",
					Operation: "REPLACE",
					Abort:     true,
				},
				{
					Source:    "var/cache/uniget/",
					Target:    cacheDirectory + "/",
					Operation: "REPLACE",
					Abort:     true,
				},
			}
			if !viper.GetBool("user") {
				logging.Debugf("Adding path rewrite rules for system installation")

				pathRewriteRules = append(pathRewriteRules, tool.PathRewrite{
					Source:    "etc/systemd/",
					Target:    "/etc/systemd/",
					Operation: "REPLACE",
				})

				if viper.GetBool("integrateprofiled") || viper.GetBool("integrateall") {
					pathRewriteRules = append(pathRewriteRules, tool.PathRewrite{
						Source:    "etc/profile.d/",
						Target:    "/etc/profile.d/",
						Operation: "REPLACE",
					})
				}

			} else {
				logging.Debugf("Adding path rewrite rules for user installation")

				pathRewriteRules = append(pathRewriteRules, tool.PathRewrite{
					Source:    "libexec/docker/cli-plugins/",
					Target:    ".docker/cli-plugins/",
					Operation: "REPLACE",
					Abort:     true,
				})

				pathRewriteRules = append(pathRewriteRules, tool.PathRewrite{
					Source:    "etc/systemd/user/",
					Target:    ".config/systemd/user/",
					Operation: "REPLACE",
					Abort:     true,
				})

				if viper.GetBool("integrateprofiled") || viper.GetBool("integrateall") {
					pathRewriteRules = append(pathRewriteRules, tool.PathRewrite{
						Source:    "etc/profile.d/",
						Target:    ".config/profile.d/",
						Operation: "REPLACE",
						Abort:     true,
					})
				}

				if viper.GetBool("integrateetc") || viper.GetBool("integrateall") {
					pathRewriteRules = append(pathRewriteRules, tool.PathRewrite{
						Source:    "etc/",
						Target:    ".config/",
						Operation: "REPLACE",
						Abort:     true,
					})
				}
			}
			if len(viper.GetString("target")) > 0 {
				targetPath := viper.GetString("target")
				if !strings.HasSuffix(targetPath, "/") {
					targetPath += "/"
				}
				pathRewriteRules = append(pathRewriteRules, tool.PathRewrite{
					Source:    "",
					Target:    targetPath,
					Operation: "PREPEND",
				})
			}
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

			switch viper.GetString("cache") {
			case "none":
				logging.Debug("Using no cache")
				toolCache = cache.NewNoneCache()

			case "file":
				logging.Debug("Using file cache")
				fileCacheDir := viper.GetString("prefix") + "/" + viper.GetString("cachedirectory")
				assertDirectory(fileCacheDir)
				toolCache = cache.NewFileCache(fileCacheDir, viper.GetInt("cacheretention"))

			case "docker":
				if containers.DockerIsAvailable() {
					logging.Debug("Using docker cache")
					toolCache, err = cache.NewDockerCache()
					if err != nil {
						return fmt.Errorf("error creating Docker cache: %s", err)
					}
				} else {
					logging.Warning.Println("Docker is not available. Falling back to no cache")
					toolCache = cache.NewNoneCache()
				}

			case "containerd":
				if containers.ContainerdIsAvailable() {
					logging.Debug("Using containerd cache")
					toolCache, err = cache.NewContainerdCache(projectName)
					if err != nil {
						return fmt.Errorf("error creating Containerd cache: %s", err)
					}
				} else {
					logging.Warning.Println("Containerd is not available. Falling back to no cache")
					toolCache = cache.NewNoneCache()
				}

			default:
				return fmt.Errorf("unsupported cache backend: %s", viper.GetString("cache"))
			}

			return nil
		},
	}
	minimumCliVersionForSchemaVersion = map[string]string{
		"1": "0.1.0",
	}
	toolCache cache.Cache = cache.NewNoneCache()
)

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
	switch altArch {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		arch = "aarch64"
	default:
		logging.Error.Printfln("Unsupported architecture: %s", arch)
		os.Exit(1)
	}

	initCacheCmd()
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
		err = viper.BindEnv(viperName, strings.ToUpper(viper.GetEnvPrefix()+"_"+strings.ReplaceAll(cobraLongName, "-", "_")))
		if err != nil {
			fmt.Printf("unable to bind environment variable for flag %s: %s", cobraLongName, err)
			os.Exit(1)
		}
	}
}

func main() {
	var err error

	viper.SetDefault("loglevel", pterm.LogLevelInfo.String())
	viper.SetDefault("debug", false)
	viper.SetDefault("trace", false)
	viper.SetDefault("prefix", "")
	viper.SetDefault("target", "usr/local")
	viper.SetDefault("user", false)
	viper.SetDefault("autoupdate", false)
	viper.SetDefault("integrateprofiled", false)
	viper.SetDefault("integrateetc", false)
	viper.SetDefault("integrateall", false)
	viper.SetDefault("registry", registry)
	viper.SetDefault("repository", imageRepository)
	viper.SetDefault("toolseparator", toolSeparator)
	viper.SetDefault("cache", "none")
	viper.SetDefault("cachedirectory", cacheDirectory+"/"+fileCacheDirectoryName)
	viper.SetDefault("cacheretention", 24*time.Hour)

	pf := rootCmd.PersistentFlags()

	pf.String("log-level", viper.GetString("loglevel"), "Log level (trace, debug, info, warning, error)")
	pf.BoolP("debug", "d", viper.GetBool("debug"), "Set log level to debug")
	pf.Bool("trace", viper.GetBool("trace"), "Set log level to trace")
	pf.StringP("prefix", "p", viper.GetString("prefix"), "Base directory for the installation (useful when preparing a chroot environment)")
	pf.StringP("target", "t", viper.GetString("target"), "Target directory for installation relative to PREFIX")
	pf.BoolP("user", "u", viper.GetBool("user"), "Install in user context")
	pf.Bool("auto-update", viper.GetBool("autoupdate"), "Automatically update metadata")
	pf.Bool("integrate-profiled", viper.GetBool("integrateprofiled"), "Integrate profile.d scripts")
	pf.Bool("integrate-etc", viper.GetBool("integrateetc"), "Integrate configuration files from /etc")
	pf.Bool("integrate-all", viper.GetBool("integrateall"), "Integrate all available integrations")
	pf.String("registry", viper.GetString("registry"), "Registry for the image repository")
	pf.String("repository", viper.GetString("repository"), "Repository for the image repository")
	pf.String("tool-separator", viper.GetString("toolseparator"), "Separator between repository and tool name")
	pf.String("cache", viper.GetString("cache"), "Cache backend to use (none, file, docker, containerd)")
	pf.String("cache-directory", viper.GetString("cachedirectory"), "Directory for the file cache")
	pf.Int("cache-retention", viper.GetInt("cacheretention"), "Retention in seconds for the file cache")
	pf.StringVar(&metadataImageTag, "metadata-image-tag", metadataImageTag, "Tag for the metadata image")

	rootCmd.MarkFlagsMutuallyExclusive("prefix", "user")
	rootCmd.MarkFlagsMutuallyExclusive("target", "user")

	err = rootCmd.Flags().MarkHidden("tool-separator")
	if err != nil {
		logging.Error.Printfln("Unable to mark tool-separator as hidden: %s", err)
		os.Exit(1)
	}
	err = rootCmd.Flags().MarkHidden("integrate-profiled")
	if err != nil {
		logging.Error.Printfln("Unable to mark integrate-profiled as hidden: %s", err)
		os.Exit(1)
	}
	err = rootCmd.Flags().MarkHidden("integrate-etc")
	if err != nil {
		logging.Error.Printfln("Unable to mark integrate-etc as hidden: %s", err)
		os.Exit(1)
	}
	err = rootCmd.Flags().MarkHidden("integrate-all")
	if err != nil {
		logging.Error.Printfln("Unable to mark integrate-all as hidden: %s", err)
		os.Exit(1)
	}
	err = rootCmd.Flags().MarkHidden("metadata-image-tag")
	if err != nil {
		logging.Error.Printfln("Unable to mark metadata-image-tag as hidden: %s", err)
		os.Exit(1)
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("uniget")

	addViperBindings(pf, "log-level", "loglevel")
	addViperBindings(pf, "debug", "debug")
	addViperBindings(pf, "trace", "trace")
	addViperBindings(pf, "prefix", "prefix")
	addViperBindings(pf, "target", "target")
	addViperBindings(pf, "user", "user")
	addViperBindings(pf, "auto-update", "autoupdate")
	addViperBindings(pf, "integrate-profiled", "integrateprofiled")
	addViperBindings(pf, "integrate-etc", "integrateetc")
	addViperBindings(pf, "integrate-all", "integrateall")
	addViperBindings(pf, "registry", "registry")
	addViperBindings(pf, "repository", "repository")
	addViperBindings(pf, "tool-separator", "toolseparator")
	addViperBindings(pf, "cache", "cache")
	addViperBindings(pf, "cache-directory", "cachedirectory")
	addViperBindings(pf, "cache-retention", "cacheretention")

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
