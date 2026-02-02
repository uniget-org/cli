package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config represents the application's configuration structure.
type Config struct {
	Server struct {
		Port    int    `mapstructure:"port"`
		Host    string `mapstructure:"host"`
		Timeout int    `mapstructure:"timeout"`
	} `mapstructure:"server"`
	Database struct {
		DSN string `mapstructure:"dsn"`
	} `mapstructure:"database"`
}

var config Config

// rootCmd is the root command for the application.
var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "A sample app using Cobra and Viper",
	Run: func(cmd *cobra.Command, args []string) {
		// Read config file
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Error reading config file: %v", err)
		}

		// Unmarshal config into struct
		if err := viper.Unmarshal(&config); err != nil {
			log.Fatalf("Unable to unmarshal config: %v", err)
		}

		// Print loaded configuration
		fmt.Printf("Server Host: %s, Port: %d\n", config.Server.Host, config.Server.Port)
		fmt.Printf("Database DSN: %s\n", config.Database.DSN)
	},
}

func init() {
	// Set config file name and type
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".") // Look for config in current directory

	// Bind environment variables to config
	viper.AutomaticEnv()

	// Optional: Bind command-line flags
	rootCmd.Flags().IntP("port", "p", 8080, "Server port")
	if err := viper.BindPFlag("server.port", rootCmd.Flags().Lookup("port")); err != nil {
		log.Fatal(err)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
