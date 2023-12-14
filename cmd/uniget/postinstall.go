package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uniget-org/cli/pkg/logging"
)

var postinstallProfileDScript = `
SCRIPTS="$( find "${target}/etc/profile.d" -type f )"
for SCRIPT in ${SCRIPTS}; do
	source "${SCRIPT}"
done
`
var postinstallCompletionScript = `
SCRIPTS="$( find "${target}/share/bash-completion/completions/" -type f )"
for SCRIPT in ${SCRIPTS}; do
	source "${SCRIPT}"
done
`

func initPostinstallCmd() {
	rootCmd.AddCommand(postinstallCmd)
}

var postinstallCmd = &cobra.Command{
	Use:   "postinstall",
	Short: "Run postinstall for tools",
	Long:  header + "\nRun postinstall for tools",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return postinstall()
	},
}

func postinstall() error {
	if directoryExists("/" + libDirectory + "/post_install") {
		entries, err := os.ReadDir("/" + libDirectory + "/post_install")
		if err != nil {
			return fmt.Errorf("unable to read post_install directory: %s", err)
		}
		scripts := make([]fs.FileInfo, 0, len(entries))
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				return fmt.Errorf("unable to get info for %s: %s", entry.Name(), err)
			}
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".sh") {
				scripts = append(scripts, info)
			}
		}
		if len(scripts) > 0 && len(viper.GetString("prefix")) > 0 {
			pterm.Warning.Printfln("prefix cannot be set for postinstall scripts to run")
			return nil
		}
		for _, file := range scripts {
			logging.Info.Printfln("Running post_install script %s", file.Name())

			logging.Debug.Printfln("Running pre_install script %s", "/"+libDirectory+"/pre_install/"+file.Name())
			cmd := exec.Command("/bin/bash", "/"+libDirectory+"/post_install/"+file.Name()) // #nosec G204 -- Tool images are a trusted source
			cmd.Env = append(os.Environ(),
				"prefix=",
				"target=/"+viper.GetString("target"),
				"arch="+arch,
				"alt_arch="+altArch,
				"uniget_contrib=/"+libDirectory+"/contrib",
			)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("unable to execute post_install script %s: %s", file.Name(), err)
			}
			fmt.Printf("%s\n", output)

			err = os.Remove("/" + libDirectory + "/post_install/" + file.Name())
			if err != nil {
				return fmt.Errorf("unable to remove post_install script %s: %s", file.Name(), err)
			}
		}
	}

	// Add shim for profile.d
	profileDShimFile := viper.GetString("prefix") + "/etc/profile.d/uniget-profile.d.sh"
	if directoryIsWritable(profileDShimFile) {
		profileDScript := strings.Replace(postinstallProfileDScript, "${target}", "/"+viper.GetString("target"), -1)
		err := os.WriteFile(
			profileDShimFile,
			[]byte(profileDScript),
			0644,
		) // #nosec G306 -- File must be executable
		if err != nil {
			return fmt.Errorf("cannot write profile.d shim: %w", err)
		}
	}

	// Add shim for completion
	completionShimFile := viper.GetString("prefix") + "/etc/profile.d/uniget-completion.sh"
	if directoryIsWritable(completionShimFile) {
		completionScript := strings.Replace(postinstallCompletionScript, "${target}", "/"+viper.GetString("target"), -1)
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
