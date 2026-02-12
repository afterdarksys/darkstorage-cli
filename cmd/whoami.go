package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current authentication status",
	Long:  `Display information about the currently authenticated user.`,
	Run: func(cmd *cobra.Command, args []string) {
		showCurrentUser()
	},
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}

func showCurrentUser() {
	apiKey := viper.GetString("api_key")
	endpoint := viper.GetString("endpoint")

	if apiKey == "" {
		color.Yellow("Not logged in")
		fmt.Println("Run 'darkstorage login' to authenticate.")
		return
	}

	// Show masked API key
	fmt.Println("Authentication Status:")
	if len(apiKey) > 12 {
		fmt.Printf("  API Key: %s...%s\n", apiKey[:8], apiKey[len(apiKey)-4:])
	} else {
		fmt.Printf("  API Key: %s\n", apiKey)
	}
	fmt.Printf("  Endpoint: %s\n", endpoint)

	// Validate token by making a request to the API
	verbose, _ := rootCmd.PersistentFlags().GetBool("verbose")
	if verbose {
		fmt.Println("\nValidating credentials...")
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", endpoint+"/v1/user/profile", nil)
	if err != nil {
		color.Red("\nError creating request: %v", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("User-Agent", "darkstorage-cli/"+version)

	resp, err := client.Do(req)
	if err != nil {
		color.Yellow("\n⚠ Unable to validate credentials: %v", err)
		fmt.Println("You may be offline or the API endpoint is unreachable.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		// Parse user info
		body, _ := io.ReadAll(resp.Body)
		var userInfo map[string]interface{}
		if err := json.Unmarshal(body, &userInfo); err == nil {
			color.Green("\n✓ Credentials valid")
			if email, ok := userInfo["email"].(string); ok {
				fmt.Printf("  User: %s\n", email)
			}
			if name, ok := userInfo["name"].(string); ok {
				fmt.Printf("  Name: %s\n", name)
			}
			if tier, ok := userInfo["tier"].(string); ok {
				fmt.Printf("  Tier: %s\n", tier)
			}
		} else {
			color.Green("\n✓ Credentials valid")
		}
	} else if resp.StatusCode == 401 {
		color.Red("\n✗ Invalid or expired credentials")
		fmt.Println("Please log in again with 'darkstorage login'")
		os.Exit(1)
	} else {
		color.Yellow("\n⚠ Unable to validate credentials (HTTP %d)", resp.StatusCode)
	}
}
