package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set configuration values",
	Long: `Set configuration values for the Dark Storage CLI.

Examples:
  darkstorage config set --key YOUR_API_KEY
  darkstorage config set --endpoint https://api.darkstorage.io
  darkstorage config set --key YOUR_KEY --endpoint https://custom.api.url`,
	Run: func(cmd *cobra.Command, args []string) {
		key, _ := cmd.Flags().GetString("key")
		endpoint, _ := cmd.Flags().GetString("endpoint")

		// Validate that at least one flag is provided
		if key == "" && endpoint == "" {
			color.Yellow("No configuration values provided.")
			fmt.Println("Use --key or --endpoint to set values.")
			fmt.Println("\nExamples:")
			fmt.Println("  darkstorage config set --key YOUR_API_KEY")
			fmt.Println("  darkstorage config set --endpoint https://api.darkstorage.io")
			os.Exit(1)
		}

		updated := []string{}

		if key != "" {
			// Basic validation for API key
			if len(key) < 20 {
				color.Yellow("Warning: API key seems too short (< 20 characters)")
			}
			viper.Set("api_key", key)
			updated = append(updated, "API key")
		}

		if endpoint != "" {
			viper.Set("endpoint", endpoint)
			updated = append(updated, "Endpoint")
		}

		// Ensure config directory exists
		home, err := os.UserHomeDir()
		if err != nil {
			color.Red("Error getting home directory: %v", err)
			os.Exit(1)
		}

		configDir := filepath.Join(home, ".darkstorage")
		if err := os.MkdirAll(configDir, 0700); err != nil {
			color.Red("Error creating config directory: %v", err)
			os.Exit(1)
		}

		configPath := filepath.Join(configDir, "config.yaml")

		if err := viper.WriteConfigAs(configPath); err != nil {
			color.Red("Error saving config: %v", err)
			os.Exit(1)
		}

		color.Green("✓ Configuration updated")
		for _, item := range updated {
			fmt.Printf("  - %s\n", item)
		}
		fmt.Printf("\nConfig file: %s\n", configPath)
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current Dark Storage CLI configuration settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		home, _ := os.UserHomeDir()
		configPath := filepath.Join(home, ".darkstorage", "config.yaml")

		fmt.Println("Current configuration:")
		fmt.Printf("  Endpoint: %s\n", viper.GetString("endpoint"))

		apiKey := viper.GetString("api_key")
		if apiKey != "" {
			if len(apiKey) > 12 {
				fmt.Printf("  API Key:  %s...%s\n", apiKey[:8], apiKey[len(apiKey)-4:])
			} else {
				fmt.Printf("  API Key:  %s\n", apiKey)
			}
		} else {
			color.Yellow("  API Key:  (not set)")
		}

		fmt.Printf("\nConfig file: %s\n", configPath)

		// Check if config file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			color.Yellow("\nNote: Config file does not exist yet. Run 'darkstorage login' or 'darkstorage config set' to create it.")
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configShowCmd)

	configSetCmd.Flags().String("key", "", "API key")
	configSetCmd.Flags().String("endpoint", "", "API endpoint")
}
