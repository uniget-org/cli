package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	var cmd = &cobra.Command{
		Use:   "viper-test",
		Short: "testing viper",
		Run: func(command *cobra.Command, args []string) {
			fmt.Printf("thing: %q\n", viper.GetString("thing"))
		},
	}

	viper.SetDefault("thing", "default")

	flags := cmd.Flags()
	flags.String("thing", viper.GetString("thing"), "The first thing")

	viper.AutomaticEnv()
	viper.BindPFlag("thing", flags.Lookup("thing"))

	cmd.Execute()
}
