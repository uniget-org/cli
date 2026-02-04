package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/pkg/semver"
)

func initSemverCmd() {
	rootCmd.AddCommand(semverCmd)
}

var semverCmd = &cobra.Command{
	Use:     "semver",
	Aliases: []string{},
	Short:   "Work with semantic versioning",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ver, _ := semver.NewSemVer("0.25.0-rc.1")
		fmt.Printf("ver: %+v\n", ver)
		fmt.Printf("major: %d\n", ver.GetMajor())
		fmt.Printf("minor: %d\n", ver.GetMinor())
		fmt.Printf("patch: %d\n", ver.GetPatch())
		fmt.Printf("prerelease tag: %s\n", ver.GetPrereleaseTag())
		fmt.Printf("prerelease: %d\n", ver.GetPrerelease())
		fmt.Printf("bump major: %s\n", ver.BumpMajor())
		fmt.Printf("bump minor: %s\n", ver.BumpMinor())
		fmt.Printf("bump patch: %s\n", ver.BumpPatch())
		fmt.Printf("bump prerelease: %s\n", ver.BumpPrerelease())

		return nil
	},
}
