package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/darkstorage/cli/internal/apikeys"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var keygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generate API keys (admin only)",
	Long: `Generate API keys for Dark Storage.

This command creates cryptographically secure API keys with customizable permissions.

Examples:
  darkstorage keygen --admin                    # Full admin key
  darkstorage keygen --name "My App"            # Standard key
  darkstorage keygen --permissions storage:*    # Storage-only key
  darkstorage keygen --read-only                # Read-only key`,
	Run: func(cmd *cobra.Command, args []string) {
		admin, _ := cmd.Flags().GetBool("admin")
		name, _ := cmd.Flags().GetString("name")
		permissions, _ := cmd.Flags().GetStringSlice("permissions")
		readOnly, _ := cmd.Flags().GetBool("read-only")
		expires, _ := cmd.Flags().GetInt("expires")
		save, _ := cmd.Flags().GetString("save")

		// Determine permissions
		var perms []string
		if admin {
			perms = []string{apikeys.PermAdminAll}
			if name == "" {
				name = "Admin Key"
			}
		} else if readOnly {
			perms = []string{apikeys.PermStorageRead}
			if name == "" {
				name = "Read-Only Key"
			}
		} else if len(permissions) > 0 {
			perms = permissions
		} else {
			// Default permissions
			perms = []string{
				apikeys.PermStorageRead,
				apikeys.PermStorageWrite,
			}
		}

		if name == "" {
			name = "Generated Key"
		}

		// Create key request
		req := &apikeys.CreateAPIKeyRequest{
			Name:          name,
			Permissions:   perms,
			ExpiresInDays: expires,
		}

		// Generate key
		generator := apikeys.NewGenerator("live")
		key, apiKeyValue, err := generator.Generate(req, "local-user")
		if err != nil {
			color.Red("✗ Failed to generate API key: %v", err)
			os.Exit(1)
		}

		// Display key info
		fmt.Println()
		color.Green("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		color.Green("  ✓ API Key Generated Successfully!")
		color.Green("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()

		color.Yellow("⚠️  SAVE THIS KEY - IT WON'T BE SHOWN AGAIN!")
		fmt.Println()

		// API Key
		color.Cyan("API Key:")
		fmt.Printf("  %s\n", apiKeyValue)
		fmt.Println()

		// S3 Credentials
		color.Cyan("S3 Credentials:")
		fmt.Printf("  Access Key: %s\n", key.S3AccessKey)
		fmt.Printf("  Secret Key: %s\n", key.S3SecretKey)
		fmt.Println()

		// Metadata
		color.Cyan("Key Information:")
		fmt.Printf("  ID:          %s\n", key.ID)
		fmt.Printf("  Name:        %s\n", key.Name)
		fmt.Printf("  Prefix:      %s...\n", key.KeyPrefix)
		fmt.Printf("  Created:     %s\n", key.CreatedAt.Format("2006-01-02 15:04:05"))
		if key.ExpiresAt != nil {
			fmt.Printf("  Expires:     %s\n", key.ExpiresAt.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("  Expires:     Never\n")
		}
		fmt.Println()

		color.Cyan("Permissions:")
		for _, perm := range key.Permissions {
			fmt.Printf("  • %s\n", perm)
		}
		fmt.Println()

		// Save to file if requested
		if save != "" {
			if err := saveKeyToFile(save, key, apiKeyValue); err != nil {
				color.Red("✗ Failed to save key: %v", err)
			} else {
				color.Green("✓ Key saved to: %s", save)
				fmt.Println()
			}
		}

		// Usage instructions
		color.Cyan("Usage:")
		fmt.Println("  1. Save the API key securely")
		fmt.Println("  2. Use it to authenticate:")
		fmt.Printf("     %s\n", color.BlueString("darkstorage login --key %s", key.KeyPrefix+"..."))
		fmt.Println()
		fmt.Println("  3. Or save to config file:")
		fmt.Printf("     %s\n", color.BlueString("echo 'api_key: %s' > ~/.darkstorage/config.yaml", apiKeyValue))
		fmt.Println()
	},
}

func saveKeyToFile(filename string, key *apikeys.APIKey, apiKeyValue string) error {
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Prepare data
	data := map[string]interface{}{
		"version": "1.0",
		"auth": map[string]string{
			"api_key":  apiKeyValue,
			"endpoint": "https://storage.darkstorage.io",
		},
		"storage": map[string]interface{}{
			"endpoint":   "storage.darkstorage.io",
			"access_key": key.S3AccessKey,
			"secret_key": key.S3SecretKey,
			"region":     "us-east-1",
			"use_ssl":    true,
		},
		"key": map[string]interface{}{
			"id":          key.ID,
			"name":        key.Name,
			"created_at":  key.CreatedAt.Format(time.RFC3339),
			"permissions": key.Permissions,
		},
	}

	// Write as JSON
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Set restrictive permissions
	if err := file.Chmod(0600); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(keygenCmd)

	keygenCmd.Flags().Bool("admin", false, "Generate admin key with all permissions")
	keygenCmd.Flags().String("name", "", "Key name/description")
	keygenCmd.Flags().StringSlice("permissions", []string{}, "Comma-separated permissions")
	keygenCmd.Flags().Bool("read-only", false, "Generate read-only key")
	keygenCmd.Flags().Int("expires", 0, "Expiration in days (0 = never)")
	keygenCmd.Flags().String("save", "", "Save key to file (e.g., ~/.darkstorage/key.json)")
}
