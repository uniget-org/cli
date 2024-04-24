package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func initDebugCmd() {
	rootCmd.AddCommand(debugCmd)
}

var debugCmd = &cobra.Command{
	Use:     "debug",
	Aliases: []string{},
	Short:   "Debug parameters",
	Long:    header + "\nDebug parameters",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("prefix: %s\n", viper.GetString("prefix"))
		fmt.Printf("target: %s\n", viper.GetString("target"))
		fmt.Printf("cacheRoot: %s\n", cacheRoot)
		fmt.Printf("cacheDirectory: %s\n", cacheDirectory)
		fmt.Printf("libRoot: %s\n", libRoot)
		fmt.Printf("libDirectory: %s\n", libDirectory)
		fmt.Printf("metadataFile: %s\n", metadataFile)

		for _, key := range viper.AllKeys() {
			fmt.Printf("viper key: %s, value: %v\n", key, viper.Get(key))
		}

		return nil
	},
}
