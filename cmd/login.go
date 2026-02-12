package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Auth Config Constants
const (
	ClientID     = "darkstorage-cli"
	RedirectPort = "4321"
	RedirectPath = "/callback"
	AuthBaseURL  = "https://console.darkstorage.io/auth/login"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to Dark Storage",
	Long: `Log in to Dark Storage using your preferred method:
  - OAuth/SSO via browser (default)
  - API key (use --key flag)

Examples:
  darkstorage login                    # OAuth/SSO login via browser
  darkstorage login --key YOUR_API_KEY # Login with API key
  darkstorage login --provider google  # Login with specific OAuth provider`,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey, _ := cmd.Flags().GetString("key")
		provider, _ := cmd.Flags().GetString("provider")

		if apiKey != "" {
			// API Key Login
			loginWithAPIKey(apiKey)
		} else {
			// OAuth/SSO Login
			performOAuthFlow(provider)
		}
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().String("key", "", "API key for direct authentication")
	loginCmd.Flags().String("provider", "afterdark", "OAuth provider (afterdark, google, github)")
}

func loginWithAPIKey(apiKey string) {
	// Validate API key format (basic check)
	if len(apiKey) < 20 {
		color.Red("Error: Invalid API key format")
		os.Exit(1)
	}

	// Save API key to config
	viper.Set("api_key", apiKey)

	home, err := os.UserHomeDir()
	if err != nil {
		color.Red("Error getting home directory: %v", err)
		os.Exit(1)
	}

	configPath := filepath.Join(home, ".darkstorage")
	if err := os.MkdirAll(configPath, 0700); err != nil {
		color.Red("Error creating config directory: %v", err)
		os.Exit(1)
	}

	configFile := filepath.Join(configPath, "config.yaml")
	if err := viper.WriteConfigAs(configFile); err != nil {
		color.Red("Error saving config: %v", err)
		os.Exit(1)
	}

	color.Green("✓ Successfully logged in with API key")
	fmt.Printf("  Config saved to: %s\n", configFile)
}

func performOAuthFlow(provider string) {
	fmt.Printf("Starting OAuth login with provider: %s\n", provider)

	// 1. Setup Local Callback Server
	tokenChan := make(chan string)
	errChan := make(chan error)

	server := &http.Server{Addr: ":" + RedirectPort}

	http.HandleFunc(RedirectPath, func(w http.ResponseWriter, r *http.Request) {
		// Check for error parameters
		if errVal := r.URL.Query().Get("error"); errVal != "" {
			errDesc := r.URL.Query().Get("error_description")
			fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head><title>Authentication Failed</title></head>
<body style="font-family: system-ui; padding: 2rem; text-align: center;">
	<h1 style="color: #dc2626;">Authentication Failed</h1>
	<p>%s</p>
	<p style="color: #666;">You can close this window and return to the terminal.</p>
</body>
</html>`, errDesc)
			errChan <- fmt.Errorf("auth error: %s - %s", errVal, errDesc)
			return
		}

		// Retrieve token
		token := r.URL.Query().Get("token")
		if token == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head><title>Authentication Failed</title></head>
<body style="font-family: system-ui; padding: 2rem; text-align: center;">
	<h1 style="color: #dc2626;">No Token Received</h1>
	<p style="color: #666;">You can close this window and return to the terminal.</p>
</body>
</html>`)
			errChan <- fmt.Errorf("no token in callback")
			return
		}

		w.Write([]byte(`
<!DOCTYPE html>
<html>
<head><title>Authentication Successful</title></head>
<body style="font-family: system-ui; padding: 2rem; text-align: center;">
	<h1 style="color: #16a34a;">✓ Authentication Successful!</h1>
	<p>You can now close this window and return to the terminal.</p>
</body>
</html>`))
		tokenChan <- token
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait a moment for server to start
	time.Sleep(100 * time.Millisecond)

	// 2. Open Browser
	redirectURL := fmt.Sprintf("http://localhost:%s%s", RedirectPort, RedirectPath)
	targetURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&provider=%s&mode=login",
		AuthBaseURL, ClientID, redirectURL, provider)

	fmt.Printf("\nOpening browser for authentication...\n")
	fmt.Printf("If the browser doesn't open, visit:\n  %s\n\n", targetURL)

	if err := browser.OpenURL(targetURL); err != nil {
		color.Yellow("Warning: Failed to open browser automatically")
		fmt.Printf("Please visit the URL above manually.\n\n")
	}

	fmt.Println("Waiting for authentication...")

	// 3. Wait for result
	select {
	case token := <-tokenChan:
		server.Shutdown(context.Background())

		fmt.Println()
		if err := saveToken(token); err != nil {
			color.Red("✗ Failed to save token: %v", err)
			os.Exit(1)
		}

		color.Green("✓ Successfully logged in!")
		home, _ := os.UserHomeDir()
		configPath := filepath.Join(home, ".darkstorage", "config.yaml")
		fmt.Printf("  Config saved to: %s\n", configPath)
		fmt.Println("\nYou can now use the Dark Storage CLI.")

	case err := <-errChan:
		server.Shutdown(context.Background())
		color.Red("\n✗ Authentication failed: %v", err)
		os.Exit(1)

	case <-time.After(5 * time.Minute):
		server.Shutdown(context.Background())
		color.Red("\n✗ Authentication timed out after 5 minutes")
		os.Exit(1)
	}
}

func saveToken(token string) error {
	viper.Set("api_key", token)

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	configPath := filepath.Join(home, ".darkstorage")
	if err := os.MkdirAll(configPath, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	configFile := filepath.Join(configPath, "config.yaml")
	if err := viper.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}
