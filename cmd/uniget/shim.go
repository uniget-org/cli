package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/uniget-org/cli/pkg/logging"
)

var profileDShim = `
SCRIPTS="$( find "${target}/etc/profile.d" -type f )"
for SCRIPT in ${SCRIPTS}; do
	source "${SCRIPT}"
done
`

func initShimCmd() {
	rootCmd.AddCommand(shimCmd)
}

var shimCmd = &cobra.Command{
	Use:     "shim",
	Aliases: []string{},
	Short:   "Install shims for profile.d",
	Long:    header + "\nInstall shims for profile.d",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := installProfileDShim()
		if err != nil {
			return fmt.Errorf("unable to install profile.d shim: %s", err)
		}

		return nil
	},
}

func installProfileDShim() error {
	profileDShimFile := profileDDirectory + "/uniget-profile.d.sh"
	profileDScript := strings.ReplaceAll(profileDShim, "${target}", "/"+viper.GetString("target"))

	if viper.GetBool("user") {
		profileDShimFile = viper.GetString("prefix") + "/.config/uniget/profile.d-shim.sh"
		profileDScript = strings.ReplaceAll(profileDShim, "${target}", viper.GetString("prefix")+"/"+viper.GetString("target"))
	}

	if fileExists(profileDShimFile) {
		file, err := os.ReadFile(profileDShimFile)
		if err != nil {
			return fmt.Errorf("cannot read profile.d shim: %w", err)
		}

		h := sha256.New()
		_, err = h.Write(file)
		if err != nil {
			return fmt.Errorf("cannot hash profile.d shim: %w", err)
		}
		fileSha256 := hex.EncodeToString(h.Sum(nil))

		h = sha256.New()
		_, err = h.Write([]byte(profileDScript))
		if err != nil {
			return fmt.Errorf("cannot hash profile.d shim: %w", err)
		}
		profileDScriptSha256 := hex.EncodeToString(h.Sum(nil))

		if fileSha256 == profileDScriptSha256 {
			logging.Info.Printfln("Profile.d shim is up to date")
			return nil
		}

		logging.Info.Printfln("Installing shim for profile.d in %s", profileDShimFile)
		if directoryIsWritable(profileDShimFile) {
			err := os.WriteFile(
				profileDShimFile,
				[]byte(profileDScript),
				0644,
			) // #nosec G306 -- File must be world-readable
			if err != nil {
				return fmt.Errorf("cannot write profile.d shim: %w", err)
			}
		}
	}

	return nil
}
