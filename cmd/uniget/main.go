package main

import (
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gitlab.com/uniget-org/cli/internal/config"

	"github.com/njayp/ophis"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"gitlab.com/uniget-org/cli/internal/constants"
	"gitlab.com/uniget-org/cli/pkg/cache"
	"gitlab.com/uniget-org/cli/pkg/containers"
	"gitlab.com/uniget-org/cli/pkg/logging"
	myos "gitlab.com/uniget-org/cli/pkg/os"
	"gitlab.com/uniget-org/cli/pkg/tool"
)

var (
	version string = "main"

	configuration *config.Config
	tools         = &tool.Tools{
		Tools: make([]tool.Tool, 0),
	}
	rootCmd = &cobra.Command{
		Use:     constants.ProjectName,
		Version: version,
		Short:   constants.Header + constants.Slogan,
		Example: `  Quickstart
    Download metadata: uniget update
    Search for tools: uniget search kubectl
    Install tools: uniget install kubectl helm`,
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			logging.OutputWriter = cmd.OutOrStdout()
			logging.ErrorWriter = cmd.ErrOrStderr()

			pterm.ThemeDefault.SuccessMessageStyle = pterm.Style{pterm.FgDefault, pterm.BgDefault}
			pterm.ThemeDefault.ErrorMessageStyle = pterm.Style{pterm.FgDefault, pterm.BgDefault}

			if !myos.IsTty() {
				pterm.DefaultSpinner.Sequence = []string{"[ ]"}
				pterm.DefaultSpinner.ShowTimer = false
			}

			if configuration.User {
				configuration.SetUserConfig()
			} else {
				configuration.SetGlobalConfig()
			}

			if configuration.Trace {
				pterm.EnableDebugMessages()
				logging.Level = pterm.LogLevelTrace

			} else if configuration.Debug {
				pterm.EnableDebugMessages()
				logging.Level = pterm.LogLevelDebug

			} else {
				pterm.DisableDebugMessages()
				logging.Level = pterm.LogLevelInfo
			}

			logging.Init()

			if len(configuration.Prefix) > 0 {
				re, err := regexp.Compile(`^\/`)
				if err != nil {
					return fmt.Errorf("cannot compile regexp: %w", err)
				}
				if !re.MatchString(configuration.Prefix) {
					wd, err := os.Getwd()
					if err != nil {
						return fmt.Errorf("cannot determine working directory: %w", err)
					}
					configuration.Prefix = wd + "/" + configuration.Prefix
					logging.Debugf("Converted prefix to absolute path %s", configuration.Prefix)
				}
			}

			if strings.HasPrefix(configuration.Target, "/") {
				configuration.Target = strings.TrimLeft(configuration.Target, "/")
			}

			if configuration.Debug {
				logging.Debugf("configuration: %s", configuration)

				logging.Debug("Path rewrite rules:")
				for _, rule := range configuration.PathRewriteRules {
					logging.Debugf("  %s -> %s (%s)", rule.Source, rule.Target, rule.Operation)
				}
			}

			if !myos.FileExists(configuration.Prefix+"/"+configuration.GetMetadataFile()) ||
				configuration.AutoUpdate ||
				(len(os.Getenv("UNIGET_IGNORE_METADATA_SIGNATURE")) > 0 &&
					!myos.FileExists(configuration.Prefix+"/"+configuration.GetMetadataFile()+".sigstore.json")) {

				logging.Debugf("Metadata does not exist. Downloading...")
				err := downloadMetadata()
				if err != nil {
					return fmt.Errorf("error downloading metadata: %s", err)
				}
			} else {
				logging.Debugf("Metadata file exists")
			}
			tools, err = loadMetadata(configuration.Prefix + "/" + configuration.GetMetadataFile())
			if err != nil {
				return fmt.Errorf("error loading metadata: %s", err)
			}

			file, err := os.Stat(configuration.Prefix + "/" + configuration.GetMetadataFile())
			if err != nil {
				return fmt.Errorf("error stating metadata file: %s", err)
			}
			now := time.Now()
			modifiedtime := file.ModTime()
			if now.Sub(modifiedtime).Hours() > 24 {
				logging.Warning.Println("Metadata file is older than 24 hours")
			}

			switch configuration.Cache {
			case "none":
				logging.Debug("Using no cache")
				toolCache = cache.NewNoneCache()

			case "file":
				logging.Debug("Using file cache")
				fileCacheDir := configuration.Prefix + "/" + configuration.FileCacheDirectoryName
				myos.AssertDirectory(fileCacheDir)
				toolCache = cache.NewFileCache(fileCacheDir, configuration.FileCacheRetention)

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
					toolCache, err = cache.NewContainerdCache(constants.ProjectName)
					if err != nil {
						return fmt.Errorf("error creating Containerd cache: %s", err)
					}
				} else {
					logging.Warning.Println("Containerd is not available. Falling back to no cache")
					toolCache = cache.NewNoneCache()
				}

			default:
				return fmt.Errorf("unsupported cache backend: %s", configuration.Cache)
			}

			return nil
		},
	}
	toolCache cache.Cache = cache.NewNoneCache()
)

func init() {
	rootCmd.AddGroup(&cobra.Group{
		ID:    "tool",
		Title: "Tool-related commands",
	})
	rootCmd.AddGroup(&cobra.Group{
		ID:    "tag",
		Title: "Tag-related commands",
	})
	rootCmd.AddGroup(&cobra.Group{
		ID:    "config",
		Title: "Configuration commands",
	})
	rootCmd.AddGroup(&cobra.Group{
		ID:    "metadata",
		Title: "Metadata commands",
	})
	rootCmd.AddGroup(&cobra.Group{
		ID:    "helper",
		Title: "Helper commands",
	})

	initBumpCmd()
	initCacheCmd()
	initCronCmd()
	initDebugCmd()
	initDescribeCmd()
	initEnvCmd()
	initGenerateCmd()
	initHealthcheckCmd()
	initHooksCmd()
	initImportCmd()
	initInspectCmd()
	initInstallCmd()
	initListCmd()
	initManpagesCmd()
	initMessageCmd()
	initRegCmd()
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

func main() {
	var err error

	configuration = config.NewDefaultConfig()

	pf := rootCmd.PersistentFlags()
	pf.StringVar(&configuration.LogLevel, "log-level", configuration.LogLevel, "Log level (trace, debug, info, warning, error)")
	pf.BoolVarP(&configuration.Debug, "debug", "d", configuration.Debug, "Set log level to debug")
	pf.BoolVar(&configuration.Trace, "trace", configuration.Trace, "Set log level to trace")
	pf.StringVarP(&configuration.Prefix, "prefix", "p", configuration.Prefix, "Base directory for the installation (useful when preparing a chroot environment)")
	pf.StringVarP(&configuration.Target, "target", "t", configuration.Target, "Target directory for installation relative to PREFIX")
	pf.BoolVarP(&configuration.User, "user", "u", configuration.User, "Install in user context")
	pf.BoolVar(&configuration.AutoUpdate, "auto-update", configuration.AutoUpdate, "Automatically update metadata")
	pf.BoolVar(&configuration.IntegrateProfileD, "integrate-profiled", configuration.IntegrateProfileD, "Integrate profile.d scripts")
	pf.BoolVar(&configuration.IntegrateEtc, "integrate-etc", configuration.IntegrateEtc, "Integrate configuration files from /etc")
	pf.BoolVar(&configuration.IntegrateAll, "integrate-all", configuration.IntegrateAll, "Integrate all available integrations")
	pf.StringVar(&configuration.Cache, "cache", configuration.Cache, "Cache backend to use (none, file, docker, containerd)")
	pf.StringVar(&configuration.FileCacheDirectoryName, "cache-directory", configuration.FileCacheDirectoryName, "Directory for the file cache")
	pf.IntVar(&configuration.FileCacheRetention, "cache-retention", configuration.FileCacheRetention, "Retention in seconds for the file cache")

	rootCmd.MarkFlagsMutuallyExclusive("prefix", "user")
	rootCmd.MarkFlagsMutuallyExclusive("target", "user")

	_ = rootCmd.Flags().MarkHidden("log-level")
	_ = rootCmd.Flags().MarkHidden("trace")
	_ = rootCmd.Flags().MarkHidden("integrate-profiled")
	_ = rootCmd.Flags().MarkHidden("integrate-etc")
	_ = rootCmd.Flags().MarkHidden("integrate-all")
	_ = rootCmd.Flags().MarkHidden("cache")
	_ = rootCmd.Flags().MarkHidden("cache-directory")
	_ = rootCmd.Flags().MarkHidden("cache-retention")

	rootCmd.SetHelpCommand(&cobra.Command{GroupID: "helper"})
	rootCmd.SetCompletionCommandGroupID("config")

	mcp := ophis.Command(nil)
	mcp.GroupID = "helper"
	rootCmd.AddCommand(mcp)

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
