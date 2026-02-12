package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out from Dark Storage",
	Long:  `Remove stored authentication credentials from your local configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		performLogout()
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func performLogout() {
	// Check if user is logged in
	apiKey := viper.GetString("api_key")
	if apiKey == "" {
		color.Yellow("You are not currently logged in.")
		return
	}

	// Clear the API key
	viper.Set("api_key", "")

	// Save the updated config
	home, err := os.UserHomeDir()
	if err != nil {
		color.Red("Error getting home directory: %v", err)
		os.Exit(1)
	}

	configPath := filepath.Join(home, ".darkstorage")
	configFile := filepath.Join(configPath, "config.yaml")

	if err := viper.WriteConfigAs(configFile); err != nil {
		color.Red("Error saving config: %v", err)
		os.Exit(1)
	}

	color.Green("✓ Successfully logged out")
	fmt.Println("  All authentication credentials have been removed.")
}
