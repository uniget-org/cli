package main

import (
	"os"

	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"
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
		Use:         "docker-setup",
		Version:     version,
		Short:       header + "The container tools installer and updater",
		SilenceUsage: true,
	}
)

func init() {
	initDockerSetup()

	initDescribeCmd()
	initInstallCmd()
	initListCmd()
	initSearchCmd()
	initTagsCmd()
	initInspectCmd()
	initUninstallCmd()
	initUpdateCmd()
	initVersionCmd()

	// TODO: Add new subcommands for executables docker-setup-<subcommand>
	//       - build
	//       - build-flat
	//       - install-from-registry
	//       - install-from-image
	//       - install-from-image-build
	//       - lego
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
		return nil
	}
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", log.WarnLevel.String(), "Log level (trace, debug, info, warning, error)")
	// TODO: Add flags --trace and --debug (make mutually exclusive)
	rootCmd.PersistentFlags().StringVarP(&prefix, "prefix", "p", "/", "Prefix for installation")
	rootCmd.PersistentFlags().StringVarP(&target, "target", "t", "usr/local", "Target directory for installation")
	rootCmd.PersistentFlags().StringVarP(&cacheDirectory, "cache-directory", "C", "/var/cache/docker-setup", "Cache directory")
	rootCmd.PersistentFlags().StringVarP(&libDirectory, "lib-directory", "L", "/var/lib/docker-setup", "Library directory")
	rootCmd.PersistentFlags().StringVarP(&metadataFileName, "metadata-file", "f", "metadata.json", "Metadata file")

	rootCmd.Execute()
}
