package main

import (
	_ "embed"
	"fmt"
	"os"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/pkg/cache"
	"gitlab.com/uniget-org/cli/pkg/logging"
	myos "gitlab.com/uniget-org/cli/pkg/os"
	"gitlab.com/uniget-org/cli/pkg/tool"
)

var (
	projectName        = "uniget"
	version     string = "main"

	//go:embed logo.txt
	header string
	slogan string = "The universal installer and updater for (container) tools" + "\n" +
		"                                       https://uniget.dev"

	config = DefaultConfig()
	tools  = tool.Tools{
		Tools: make([]tool.Tool, 0),
	}
	rootCmd = &cobra.Command{
		Use:          projectName,
		Version:      version,
		Short:        header + slogan,
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logging.OutputWriter = cmd.OutOrStdout()
			logging.ErrorWriter = cmd.ErrOrStderr()

			pterm.ThemeDefault.SuccessMessageStyle = pterm.Style{pterm.FgDefault, pterm.BgDefault}
			pterm.ThemeDefault.ErrorMessageStyle = pterm.Style{pterm.FgDefault, pterm.BgDefault}

			if !myos.IsTty() {
				pterm.DefaultSpinner.Sequence = []string{"[ ]"}
				pterm.DefaultSpinner.ShowTimer = false
			}

			if config.trace {
				pterm.EnableDebugMessages()
				logging.Level = pterm.LogLevelTrace

			} else if config.debug {
				pterm.EnableDebugMessages()
				logging.Level = pterm.LogLevelDebug

			} else {
				pterm.DisableDebugMessages()
				logging.Level = pterm.LogLevelInfo
			}

			logging.Init()
			err := config.Update()
			if err != nil {
				return fmt.Errorf("error updating config: %w", err)
			}

			if config.debug {
				config.Debug()
			}

			if !fileExists(config.prefix + "/" + config.metadataFile) {
				logging.Debugf("Metadata file does not exist. Downloading...")
				err := downloadMetadata()
				if err != nil {
					return fmt.Errorf("error downloading metadata: %s", err)
				}
			} else {
				logging.Debugf("Metadata file exists")
			}
			err = loadMetadata()
			if err != nil {
				return fmt.Errorf("error loading metadata: %s", err)
			}

			file, err := os.Stat(config.prefix + "/" + config.metadataFile)
			if err != nil {
				return fmt.Errorf("error stating metadata file: %s", err)
			}
			now := time.Now()
			modifiedtime := file.ModTime()
			if now.Sub(modifiedtime).Hours() > 24 {
				logging.Warning.Println("Metadata file is older than 24 hours")
			}

			return nil
		},
	}
	minimumCliVersionForSchemaVersion = map[string]string{
		"1": "0.1.0",
	}
	toolCache cache.Cache = cache.NewNoneCache()
)

func init() {
	initBumpCmd()
	initCacheCmd()
	initCronCmd()
	initDebugCmd()
	initDescribeCmd()
	initEnvCmd()
	initGenerateCmd()
	initHealthcheckCmd()
	initHooksCmd()
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

	pf := rootCmd.PersistentFlags()

	pf.StringVar(&config.logLevel, "log-level", config.logLevel, "Log level (trace, debug, info, warning, error)")
	pf.BoolVarP(&config.debug, "debug", "d", config.debug, "Set log level to debug")
	pf.BoolVar(&config.trace, "trace", config.trace, "Set log level to trace")
	pf.StringVarP(&config.prefix, "prefix", "p", config.prefix, "Base directory for the installation (useful when preparing a chroot environment)")
	pf.StringVarP(&config.target, "target", "t", config.target, "Target directory for installation relative to PREFIX")
	pf.BoolVarP(&config.user, "user", "u", config.user, "Install in user context")
	pf.BoolVar(&config.autoupdate, "auto-update", config.autoupdate, "Automatically update metadata")
	pf.BoolVar(&config.integrateprofiled, "integrate-profiled", config.integrateprofiled, "Integrate profile.d scripts")
	pf.BoolVar(&config.integrateetc, "integrate-etc", config.integrateetc, "Integrate configuration files from /etc")
	pf.BoolVar(&config.integrateall, "integrate-all", config.integrateall, "Integrate all available integrations")
	pf.StringVar(&config.registry, "registry", config.registry, "Registry for the image repository")
	pf.StringVar(&config.imageRepository, "repository", config.imageRepository, "Repository for the image repository")
	pf.StringVar(&config.toolSeparator, "tool-separator", config.toolSeparator, "Separator between repository and tool name")
	pf.StringVar(&config.cache, "cache", config.cache, "Cache backend to use (none, file, docker, containerd)")
	pf.StringVar(&config.cacheDirectory, "cache-directory", config.cachedirectory, "Directory for the file cache")
	pf.IntVar(&config.cacheretention, "cache-retention", config.cacheretention, "Retention in seconds for the file cache")
	pf.StringVar(&config.metadataImageTag, "metadata-image-tag", config.metadataImageTag, "Tag for the metadata image")

	rootCmd.MarkFlagsMutuallyExclusive("prefix", "user")
	rootCmd.MarkFlagsMutuallyExclusive("target", "user")

	err = rootCmd.Flags().MarkHidden("registry")
	if err != nil {
		logging.Error.Printfln("Unable to mark registry as hidden: %s", err)
		os.Exit(1)
	}
	err = rootCmd.Flags().MarkHidden("repository")
	if err != nil {
		logging.Error.Printfln("Unable to mark repository as hidden: %s", err)
		os.Exit(1)
	}
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

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
