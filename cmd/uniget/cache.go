package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/moby/moby/client"
	"gitlab.com/uniget-org/cli/pkg/containers"
	myos "gitlab.com/uniget-org/cli/pkg/os"
)

func initCacheCmd() {
	rootCmd.AddCommand(cacheCmd)

	cacheCmd.AddCommand(cacheInfoCmd)
	cacheCmd.AddCommand(cacheStatsCmd)
	cacheCmd.AddCommand(cacheListCmd)
	cacheCmd.AddCommand(cachePruneCmd)
}

var cacheCmd = &cobra.Command{
	Use: "cache",
	Aliases: []string{
		"c",
	},
	Short: "Manage the cache",
	Long:  header + "\nManage the cache",
	Args:  cobra.NoArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := rootCmd.PersistentPreRunE(cmd, args)
		if err != nil {
			return err
		}
		if !cacheIsConfigured() {
			return fmt.Errorf("cache is not configured")
		}
		return nil
	},
}

var cacheInfoCmd = &cobra.Command{
	Use: "info",
	Aliases: []string{
		"i",
	},
	Short: "Display cache information",
	Long:  header + "\nDisplay cache information",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Cache type     : %s\n", viper.GetString("cache"))
		fmt.Printf("Cache retention: %s\n", viper.GetString("cacheretention"))
		if viper.GetString("cache") == "file" {
			fmt.Printf("Cache directory: %s\n", viper.GetString("prefix")+"/"+viper.GetString("cachedirectory"))
		}
		return nil
	},
}

var cacheStatsCmd = &cobra.Command{
	Use: "stats",
	Aliases: []string{
		"s",
	},
	Short: "Display cache statistics",
	Long:  header + "\nDisplay cache statistics",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var size int64
		var err error

		switch viper.GetString("cache") {
		case "file":
			if viper.GetString("cache") == "file" {
				fmt.Printf("Cache directory: %s\n", viper.GetString("prefix")+"/"+viper.GetString("cachedirectory"))
			}
			size, err = dirSize(
				viper.GetString("prefix") + "/" + viper.GetString("cachedirectory"),
			)
			if err != nil {
				return fmt.Errorf("error calculating cache size: %v", err)
			}

		case "docker":
			cli, err := client.New(client.FromEnv)
			if err != nil {
				return fmt.Errorf("failed to create Docker client: %w", err)
			}
			images, err := containers.ListDockerImagesByPrefix(cli, "ghcr.io/uniget-org/tools/")
			if err != nil {
				return fmt.Errorf("failed to list Docker images: %w", err)
			}
			for _, img := range images {
				if img.RepoTags == nil {
					continue
				}
				size += img.Size
			}

		case "containerd":
			return fmt.Errorf("cache type 'containerd' is not yet implemented")
		}

		fmt.Printf("Cache size     : %s\n", myos.ConvertBytesToHumanReadable(size))
		return nil
	},
}

var cacheListCmd = &cobra.Command{
	Use: "list",
	Aliases: []string{
		"l",
		"ls",
	},
	Short: "Display cache contents",
	Long:  header + "\nDisplay cache contents",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		switch viper.GetString("cache") {
		case "file":
			if viper.GetString("cache") == "file" {
				fmt.Printf("Cache directory: %s\n", viper.GetString("prefix")+"/"+viper.GetString("cachedirectory"))
			}
			files, err := dirList(
				viper.GetString("prefix") + "/" + viper.GetString("cachedirectory"),
			)
			if err != nil {
				return fmt.Errorf("error calculating cache size: %v", err)
			}
			for _, file := range files {
				fmt.Println(file)
			}

		case "docker":
			cli, err := client.New(client.FromEnv)
			if err != nil {
				return fmt.Errorf("failed to create Docker client: %w", err)
			}
			images, err := containers.ListDockerImagesByPrefix(cli, "ghcr.io/uniget-org/tools/")
			if err != nil {
				return fmt.Errorf("failed to list Docker images: %w", err)
			}
			for _, img := range images {
				if img.RepoTags == nil {
					continue
				}
				for _, tag := range img.RepoTags {
					fmt.Println(tag)
				}
			}

		case "containerd":
			return fmt.Errorf("cache type 'containerd' is not yet implemented")
		}

		return nil
	},
}

var cachePruneCmd = &cobra.Command{
	Use: "prune",
	Aliases: []string{
		"p",
	},
	Short: "Remove unused cache entries",
	Long:  header + "\nRemove unused cache entries",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var count int
		var err error

		switch viper.GetString("cache") {
		case "file":
			if viper.GetString("cache") == "file" {
				fmt.Printf("Cache directory: %s\n", viper.GetString("prefix")+"/"+viper.GetString("cachedirectory"))
			}
			count, err = dirPrune(
				viper.GetString("prefix") + "/" + viper.GetString("cachedirectory"),
			)
			if err != nil {
				return fmt.Errorf("error pruning cache: %v", err)
			}

		case "docker":
			cli, err := client.New(client.FromEnv)
			if err != nil {
				return fmt.Errorf("failed to create Docker client: %w", err)
			}
			images, err := containers.ListDockerImagesByPrefix(cli, "ghcr.io/uniget-org/tools/")
			if err != nil {
				return fmt.Errorf("failed to list Docker images: %w", err)
			}
			for _, img := range images {
				if img.RepoTags == nil {
					continue
				}
				for _, tag := range img.RepoTags {
					if err := containers.RemoveDockerImage(cli, tag); err != nil {
						fmt.Fprintf(os.Stderr, "Failed to remove image %s: %v\n", tag, err)
					} else {
						count++
					}
				}
			}

		case "containerd":
			return fmt.Errorf("cache type 'containerd' is not yet implemented")
		}

		fmt.Printf("Removed %d cache entries.\n", count)
		return nil
	},
}

func cacheIsConfigured() bool {
	if viper.GetString("cache") == "" || viper.GetString("cache") == "none" {
		return false
	}
	return true
}

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func dirList(path string) ([]string, error) {
	var files []string
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, info.Name())
		}
		return err
	})
	return files, err
}

func dirPrune(path string) (int, error) {
	var deleted int
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if removeErr := os.Remove(p); removeErr != nil {
				return fmt.Errorf("failed to remove file %s: %w", p, removeErr)
			}
			deleted++
		}
		return nil
	})
	return deleted, err
}
