package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/regclient/regclient/pkg/template"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/platform"
	"github.com/regclient/regclient/types/ref"
	"github.com/spf13/cobra"
	"github.com/uniget-org/cli/pkg/containers"
)

var (
	regVersion          = "main"
	regFormat           = "pretty"
	regManifestPlatform = ""
	regSizeHuman        = false
)

func initRegCmd() {
	regIndexCmd.Flags().StringVarP(&regVersion, "version", "", regVersion, "Specify the version for the index")
	regIndexCmd.Flags().StringVarP(&regFormat, "format", "f", regFormat, "Specify the output format (pretty|json)")

	regManifestCmd.Flags().StringVarP(&regVersion, "version", "", regVersion, "Specify the version for the manifest")
	regManifestCmd.Flags().StringVarP(&regManifestPlatform, "platform", "", "", "Specify the platform for the manifest")
	regManifestCmd.Flags().StringVarP(&regFormat, "format", "f", regFormat, "Specify the output format (pretty|json)")

	regSizeCmd.Flags().StringVarP(&regVersion, "version", "", regVersion, "Specify the version for the manifest")
	regSizeCmd.Flags().StringVarP(&regManifestPlatform, "platform", "", "", "Specify the platform for the manifest")
	regSizeCmd.Flags().BoolVarP(&regSizeHuman, "human", "H", regSizeHuman, "Display size in human-readable format")
	regSizeCmd.Flags().StringVarP(&regFormat, "format", "f", regFormat, "Specify the output format (json|text)")

	regRefCmd.Flags().StringVarP(&regVersion, "version", "", regVersion, "Specify the version for the reference")

	regCmd.AddCommand(regIndexCmd)
	regCmd.AddCommand(regManifestCmd)
	regCmd.AddCommand(regSizeCmd)
	regCmd.AddCommand(regTagsCmd)
	regCmd.AddCommand(regRefCmd)

	rootCmd.AddCommand(regCmd)
}

var regCmd = &cobra.Command{
	Use:     "registry",
	Aliases: []string{"reg", "r"},
	Short:   "Display installation paths as environment variables",
	Long:    header + "\nDisplay installation paths as environment variables",
	Hidden:  true,
}

func getFormatString() string {
	format := "{{printPretty .}}"
	switch regFormat {
	case "pretty":
		format = "{{printPretty .}}"
	case "json":
		format = "{{printf \"%s\" .RawBody}}"
	}

	return format
}

func buildReference(tool string) string {
	return registry + "/" + imageRepository + toolSeparator + tool + ":" + regVersion
}

var regRefCmd = &cobra.Command{
	Use:     "reference",
	Aliases: []string{"ref", "r"},
	Short:   "Display image reference",
	Long:    header + "\nDisplay image reference",
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(buildReference(args[0]))

		return nil
	},
}

var regIndexCmd = &cobra.Command{
	Use:     "index",
	Aliases: []string{"i"},
	Short:   "Display image index",
	Long:    header + "\nDisplay image index",
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		format := getFormatString()

		ctx := context.Background()

		image := buildReference(args[0])
		r, err := ref.New(image)
		if err != nil {
			return fmt.Errorf("failed to parse image name <%s>: %s", image, err)
		}

		rc := containers.GetRegclient()
		//nolint:errcheck
		defer rc.Close(ctx, r)

		m, err := rc.ManifestGet(ctx, r)
		if err != nil {
			return fmt.Errorf("failed to get manifest: %s", err)
		}

		err = template.Writer(cmd.OutOrStdout(), format, m)
		if err != nil {
			return fmt.Errorf("failed to write template: %s", err)
		}

		return nil
	},
}

var regManifestCmd = &cobra.Command{
	Use:     "manifest",
	Aliases: []string{"m"},
	Short:   "Display image manifest",
	Long:    header + "\nDisplay image manifest",
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		format := getFormatString()

		ctx := context.Background()

		image := buildReference(args[0])
		r, err := ref.New(image)
		if err != nil {
			return fmt.Errorf("failed to parse image name <%s>: %s", image, err)
		}

		rc := containers.GetRegclient()
		//nolint:errcheck
		defer rc.Close(ctx, r)

		if regManifestPlatform == "" {
			regManifestPlatform = platform.Local().String()
		}

		p, err := platform.Parse(regManifestPlatform)
		if err != nil {
			return fmt.Errorf("failed to parse platform <%s>: %s", regManifestPlatform, err)
		}
		m, err := containers.GetPlatformManifest(ctx, rc, r, p)
		if err != nil {
			return fmt.Errorf("failed to get platform manifest: %s", err)
		}

		err = template.Writer(cmd.OutOrStdout(), format, m)
		if err != nil {
			return fmt.Errorf("failed to write template: %s", err)
		}

		return nil
	},
}

var regSizeCmd = &cobra.Command{
	Use:     "size",
	Aliases: []string{"s"},
	Short:   "Display size of first layer",
	Long:    header + "\nDisplay image size",
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		image := buildReference(args[0])
		r, err := ref.New(image)
		if err != nil {
			return fmt.Errorf("failed to parse image name <%s>: %s", image, err)
		}

		rc := containers.GetRegclient()
		//nolint:errcheck
		defer rc.Close(ctx, r)

		if regManifestPlatform == "" {
			regManifestPlatform = platform.Local().String()
		}

		p, err := platform.Parse(regManifestPlatform)
		if err != nil {
			return fmt.Errorf("failed to parse platform <%s>: %s", regManifestPlatform, err)
		}
		m, err := containers.GetPlatformManifest(ctx, rc, r, p)
		if err != nil {
			return fmt.Errorf("failed to get platform manifest: %s", err)
		}

		mi, ok := m.(manifest.Imager)
		if !ok {
			return fmt.Errorf("failed to assert manifest as imager: %s", err)
		}
		size, err := mi.GetSize()
		if err != nil {
			return fmt.Errorf("failed to get size: %s", err)
		}

		units := []string{"B", "KB", "MB", "GB", "TB"}
		index := 0
		if regSizeHuman {
			for size > 1024 {
				size /= 1024
				index += 1
			}
			fmt.Printf("%v%s\n", size, units[index])

		} else {
			fmt.Printf("%v\n", size)
		}

		return nil
	},
}

var regTagsCmd = &cobra.Command{
	Use:     "tags",
	Aliases: []string{"t"},
	Short:   "Display tags for image",
	Long:    header + "\nDisplay tags for image",
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		image := buildReference(args[0])
		r, err := ref.New(image)
		if err != nil {
			return fmt.Errorf("failed to parse image name <%s>: %s", image, err)
		}

		rc := containers.GetRegclient()
		//nolint:errcheck
		defer rc.Close(ctx, r)

		tags, err := rc.TagList(ctx, r)
		if err != nil {
			return fmt.Errorf("failed to list tags: %s", err)
		}

		sortedTags := tags.Tags
		sort.Strings(sortedTags)
		for _, tag := range sortedTags {
			fmt.Println(tag)
		}

		return nil
	},
}
