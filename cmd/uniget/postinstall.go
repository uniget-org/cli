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
		err := postinstall()
		if err != nil {
			return fmt.Errorf("unable to run postinstall: %s", err)
		}

		err = installProfileDShim()
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

			logging.Debugf("Running post_install script %s", "/"+libDirectory+"/post_install/"+file.Name())
			cmd := exec.Command("/bin/bash", "/"+libDirectory+"/post_install/"+file.Name()) // #nosec G204 -- Tool images are a trusted source
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, "target=/"+viper.GetString("target"))
			cmd.Env = append(cmd.Env, "arch="+arch)
			cmd.Env = append(cmd.Env, "alt_arch="+altArch)
			cmd.Env = append(cmd.Env, "uniget_contrib=/"+libDirectory+"/contrib")
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Print("---------- 8< ----------\n")
				fmt.Printf("%s\n", output)
				fmt.Print("---------- 8< ----------\n")
				return fmt.Errorf("unable to execute post_install script %s: %s", file.Name(), err)
			}
			fmt.Printf("%s\n", output)

			err = os.Remove("/" + libDirectory + "/post_install/" + file.Name())
			if err != nil {
				return fmt.Errorf("unable to remove post_install script %s: %s", file.Name(), err)
			}
		}
	}

	return nil
}

func installProfileDShim() error {
	profileDShimFile := profileDDirectory + "/uniget-profile.d.sh"
	profileDScript := strings.Replace(postinstallProfileDScript, "${target}", "/"+viper.GetString("target"), -1)

	if viper.GetBool("user") {
		profileDShimFile = viper.GetString("prefix") + "/.config/uniget/profile.d-shim.sh"
		profileDScript = strings.Replace(postinstallProfileDScript, "${target}", viper.GetString("prefix")+"/"+viper.GetString("target"), -1)
	}

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
	completionScript := strings.Replace(postinstallCompletionScript, "${target}", "/"+viper.GetString("target"), -1)

	if viper.GetBool("user") {
		dataDirectory := ".local/share"
		if os.Getenv("XDG_DATA_HOME") != "" {
			if strings.HasPrefix(os.Getenv("XDG_DATA_HOME"), os.Getenv("HOME")) {
				dataDirectory = strings.TrimPrefix(os.Getenv("XDG_DATA_HOME"), os.Getenv("HOME")+"/")
			}
		}
		completionShimFile = viper.GetString("prefix") + dataDirectory + "/bash-completion/uniget-shim.sh"
		completionScript = strings.Replace(postinstallCompletionScript, "${target}", viper.GetString("prefix")+"/.local", -1)
	}

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

func installSystemDUnit() error {
	// add flag --systemd whether to install systemd units
	// check if tool ships with systemd unit (check file list for etc/systemd/system/*)
	// if no user context create symlink from TARGET/etc/systemd/system/* to /etc/systemd/system/
	// if user context create symlink from TARGET/etc/systemd/user/* to ~/.local/share/systemd/user/ or ~/.config/systemd/user/
	// reload systemd honoring context

	return nil
}
