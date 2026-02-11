package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/pkg/semver"
)

func initSemverCmd() {
	getSemverCmd.AddCommand(getMajorSemverCmd)
	getSemverCmd.AddCommand(getMinorSemverCmd)
	getSemverCmd.AddCommand(getPatchSemverCmd)
	getSemverCmd.AddCommand(getPrereleaseSemverCmd)
	semverCmd.AddCommand(getSemverCmd)

	bumpSemverCmd.AddCommand(bumpMajorSemverCmd)
	bumpSemverCmd.AddCommand(bumpMinorSemverCmd)
	bumpSemverCmd.AddCommand(bumpPatchSemverCmd)
	bumpSemverCmd.AddCommand(bumpPrereleaseSemverCmd)
	semverCmd.AddCommand(bumpSemverCmd)

	rootCmd.AddCommand(semverCmd)
}

var semverCmd = &cobra.Command{
	Use:     "semver",
	Aliases: []string{},
	Short:   "Work with semantic versioning",
}

var getSemverCmd = &cobra.Command{
	Use: "get",
	Aliases: []string{
		"g",
	},
	Short: "Get part of version",
}

var getMajorSemverCmd = &cobra.Command{
	Use:     "major",
	Aliases: []string{},
	Short:   "Get major version",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ver, _ := semver.NewSemVer(args[0])
		fmt.Printf("%d\n", ver.GetMajor())

		return nil
	},
}

var getMinorSemverCmd = &cobra.Command{
	Use:     "minor",
	Aliases: []string{},
	Short:   "Get minor version",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ver, _ := semver.NewSemVer(args[0])
		fmt.Printf("%d\n", ver.GetMinor())

		return nil
	},
}

var getPatchSemverCmd = &cobra.Command{
	Use:     "patch",
	Aliases: []string{},
	Short:   "Get patch version",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ver, _ := semver.NewSemVer(args[0])
		fmt.Printf("%d\n", ver.GetPatch())

		return nil
	},
}

var getPrereleaseSemverCmd = &cobra.Command{
	Use:     "prerelease",
	Aliases: []string{},
	Short:   "Get prerelease version",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ver, _ := semver.NewSemVer(args[0])
		fmt.Printf("%s.%d\n", ver.GetPrereleaseTag(), ver.GetPrerelease())

		return nil
	},
}

var bumpSemverCmd = &cobra.Command{
	Use: "bump",
	Aliases: []string{
		"b",
	},
	Short: "Bump version",
}

var bumpMajorSemverCmd = &cobra.Command{
	Use:     "major",
	Aliases: []string{},
	Short:   "Bump major version",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ver, _ := semver.NewSemVer(args[0])
		fmt.Printf("%s\n", ver.BumpMajor())

		return nil
	},
}

var bumpMinorSemverCmd = &cobra.Command{
	Use:     "minor",
	Aliases: []string{},
	Short:   "Bump minor version",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ver, _ := semver.NewSemVer(args[0])
		fmt.Printf("%s\n", ver.BumpMinor())

		return nil
	},
}

var bumpPatchSemverCmd = &cobra.Command{
	Use:     "patch",
	Aliases: []string{},
	Short:   "Bump patch version",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ver, _ := semver.NewSemVer(args[0])
		fmt.Printf("%s\n", ver.BumpPatch())

		return nil
	},
}

var bumpPrereleaseSemverCmd = &cobra.Command{
	Use:     "prerelease",
	Aliases: []string{},
	Short:   "Bump prerelease version",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ver, _ := semver.NewSemVer(args[0])
		fmt.Printf("%s\n", ver.BumpPrerelease())

		return nil
	},
}
