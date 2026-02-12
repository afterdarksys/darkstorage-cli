package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	version string
)

var rootCmd = &cobra.Command{
	Use:   "darkstorage",
	Short: "Dark Storage CLI - API-first cloud storage",
	Long: `Dark Storage CLI provides command-line access to all Dark Storage features.

Upload, download, and manage your files with ease. Full S3-compatible
operations, block storage management, and security features.

Get started:
  darkstorage login
  darkstorage ls
  darkstorage put ./file.txt my-bucket/`,
}

func Execute() error {
	return rootCmd.Execute()
}

func SetVersion(v string) {
	version = v
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.darkstorage/config.yaml)")
	rootCmd.PersistentFlags().String("api-key", "", "API key (overrides config)")
	rootCmd.PersistentFlags().String("endpoint", "https://api.darkstorage.io", "API endpoint")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().Bool("json", false, "output in JSON format")

	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("endpoint", rootCmd.PersistentFlags().Lookup("endpoint"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		configPath := filepath.Join(home, ".darkstorage")
		os.MkdirAll(configPath, 0700)

		viper.AddConfigPath(configPath)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("DARKSTORAGE")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// Config file not found is okay for first run
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintln(os.Stderr, "Error reading config:", err)
		}
	}
}
