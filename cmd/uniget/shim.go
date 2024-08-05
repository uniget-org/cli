package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uniget-org/cli/pkg/logging"
)

var profileDShim = `
SCRIPTS="$( find "${target}/etc/profile.d" -type f )"
for SCRIPT in ${SCRIPTS}; do
	source "${SCRIPT}"
done
`
var completionShim = `
SCRIPTS="$( find "${target}/share/bash-completion/completions/" -type f )"
for SCRIPT in ${SCRIPTS}; do
	source "${SCRIPT}"
done
`

func initShimCmd() {
	rootCmd.AddCommand(shimCmd)
}

var shimCmd = &cobra.Command{
	Use:   "shim",
	Aliases: []string{"postinstall"},
	Short: "Install shims for profile.d and completion scripts",
	Long:  header + "\nInstall shims for profile.d and completion scripts",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if cmd.CalledAs() == "postinstall" {
			logging.Warning.Println("The 'postinstall' command is deprecated and will be removed in a future release. Please use 'shim' instead.")
		}

		err := installProfileDShim()
		if err != nil {
			return fmt.Errorf("unable to install profile.d shim: %s", err)
		}

		err = installCompletionShim()
		if err != nil {
			return fmt.Errorf("unable to install completion shim: %s", err)
		}

		return nil
	},
}

func installProfileDShim() error {
	profileDShimFile := profileDDirectory + "/uniget-profile.d.sh"
	profileDScript := strings.Replace(profileDShim, "${target}", "/"+viper.GetString("target"), -1)

	if viper.GetBool("user") {
		profileDShimFile = viper.GetString("prefix") + "/.config/uniget/profile.d-shim.sh"
		profileDScript = strings.Replace(completionShim, "${target}", viper.GetString("prefix")+"/"+viper.GetString("target"), -1)
	}

	logging.Info.Printfln("Installing shim for profile.d in %s", profileDShimFile)
		
	if directoryIsWritable(profileDShimFile) {
		err := os.WriteFile(
			profileDShimFile,
			[]byte(profileDScript),
			0644,
		) // #nosec G306 -- File must be executable
		if err != nil {
			return fmt.Errorf("cannot write profile.d shim: %w", err)
		}
	}

	return nil
}

func installCompletionShim() error {
	completionShimFile := profileDDirectory + "/uniget-completion.sh"
	completionScript := strings.Replace(completionShim, "${target}", "/"+viper.GetString("target"), -1)

	if viper.GetBool("user") {
		dataDirectory := ".local/share"
		if os.Getenv("XDG_DATA_HOME") != "" {
			if strings.HasPrefix(os.Getenv("XDG_DATA_HOME"), os.Getenv("HOME")) {
				dataDirectory = strings.TrimPrefix(os.Getenv("XDG_DATA_HOME"), os.Getenv("HOME")+"/")
			}
		}
		completionShimFile = viper.GetString("prefix") + dataDirectory + "/bash-completion/uniget-shim.sh"
		completionScript = strings.Replace(completionShim, "${target}", viper.GetString("prefix")+"/.local", -1)
	}

	logging.Info.Printfln("Installing shim for completion in %s", completionShimFile)

	if directoryIsWritable(completionShimFile) {
		err := os.WriteFile(
			completionShimFile,
			[]byte(completionScript),
			0644,
		) // #nosec G306 -- File must be executable
		if err != nil {
			return fmt.Errorf("cannot write completion shim: %w", err)
		}
	}

	return nil
}
