package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func initDevCmd() {
	rootCmd.AddCommand(devCmd)
}

var devCmd = &cobra.Command{
	Use:       "dev",
	Short:     "Maintainer tools",
	Long:      header + "\nMaintainer tools",
	Args:      cobra.ExactArgs(1),
	ValidArgs: tools.GetNames(),
	RunE:      func(cmd *cobra.Command, args []string) error {
		dev := exec.Command("/bin/bash", "./scripts/dev.sh")
		dev.Env = append(os.Environ(), "TOOL="+args[0])
		output, err := dev.CombinedOutput()
		if err != nil {
			return fmt.Errorf("unable to run dev script for %s: %s", args[0], err)
		}
		fmt.Println(string(output))

		return nil
	},
}
