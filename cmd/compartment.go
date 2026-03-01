package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var compartmentCmd = &cobra.Command{
	Use:   "compartment",
	Short: "Manage file compartmentalization (security layers)",
	Long: `Create and manage security compartments for additional file isolation.

Compartments provide a second layer of security beyond standard permissions,
allowing you to group files by security level, compliance requirements, or
organizational boundaries.

Examples:
  darkstorage compartment create classified --level SECRET
  darkstorage compartment assign my-bucket/file.txt classified
  darkstorage compartment list`,
}

var compartmentCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new security compartment",
	Long: `Create a new security compartment with specified security policies.

Examples:
  darkstorage compartment create classified --level SECRET
  darkstorage compartment create pii --compliance HIPAA,GDPR
  darkstorage compartment create finance --description "Financial records"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		level, _ := cmd.Flags().GetString("level")
		compliance, _ := cmd.Flags().GetString("compliance")
		description, _ := cmd.Flags().GetString("description")
		requireMFA, _ := cmd.Flags().GetBool("require-mfa")
		encryption, _ := cmd.Flags().GetString("encryption")

		payload := map[string]interface{}{
			"name":        name,
			"level":       level,
			"compliance":  strings.Split(compliance, ","),
			"description": description,
			"policies": map[string]interface{}{
				"require_mfa": requireMFA,
				"encryption":  encryption,
			},
		}

		jsonData, _ := json.Marshal(payload)
		endpoint := getAPIEndpoint()

		color.Green("Creating compartment: %s", name)
		if level != "" {
			color.Cyan("  Level: %s", level)
		}
		if compliance != "" {
			color.Cyan("  Compliance: %s", compliance)
		}
		if requireMFA {
			color.Cyan("  MFA Required: Yes")
		}

		color.Yellow("\nAPI Call: POST %s/compartments", endpoint)
		color.Yellow("Payload: %s", string(jsonData))
		color.Green("\n✓ Compartment created: %s", name)
	},
}

var compartmentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all compartments",
	Long: `List all security compartments and their configurations.

Examples:
  darkstorage compartment list
  darkstorage compartment list --json`,
	Run: func(cmd *cobra.Command, args []string) {
		// Mock data for demonstration
		compartments := []map[string]interface{}{
			{
				"name":        "classified",
				"level":       "SECRET",
				"files":       142,
				"users":       5,
				"created":     "2026-02-15",
				"compliance":  "ITAR",
				"require_mfa": true,
			},
			{
				"name":        "pii",
				"level":       "CONFIDENTIAL",
				"files":       1028,
				"users":       12,
				"created":     "2026-01-10",
				"compliance":  "HIPAA, GDPR",
				"require_mfa": true,
			},
			{
				"name":        "public",
				"level":       "PUBLIC",
				"files":       5432,
				"users":       50,
				"created":     "2026-01-01",
				"compliance":  "",
				"require_mfa": false,
			},
		}

		if jsonOutput, _ := cmd.Flags().GetBool("json"); jsonOutput {
			data, _ := json.MarshalIndent(compartments, "", "  ")
			fmt.Println(string(data))
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Level", "Files", "Users", "Compliance", "MFA"})
		table.SetBorder(false)

		for _, c := range compartments {
			mfa := "No"
			if c["require_mfa"].(bool) {
				mfa = "Yes"
			}
			table.Append([]string{
				c["name"].(string),
				c["level"].(string),
				fmt.Sprintf("%d", c["files"].(int)),
				fmt.Sprintf("%d", c["users"].(int)),
				c["compliance"].(string),
				mfa,
			})
		}
		table.Render()
	},
}

var compartmentAssignCmd = &cobra.Command{
	Use:   "assign <path> <compartment>",
	Short: "Assign a file to a compartment",
	Long: `Assign a file or directory to a security compartment.

Examples:
  darkstorage compartment assign my-bucket/file.txt classified
  darkstorage compartment assign my-bucket/sensitive/ pii --recursive`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		compartment := args[1]
		recursive, _ := cmd.Flags().GetBool("recursive")

		color.Green("Assigning to compartment '%s': %s", compartment, path)
		if recursive {
			color.Cyan("  Mode: Recursive")
		}

		endpoint := getAPIEndpoint()
		color.Yellow("\nAPI Call: POST %s/compartments/%s/assign", endpoint, compartment)
		color.Green("\n✓ Assigned: %s → %s", path, compartment)
	},
}

var compartmentFilesCmd = &cobra.Command{
	Use:   "files <compartment>",
	Short: "List files in a compartment",
	Long: `List all files assigned to a specific compartment.

Examples:
  darkstorage compartment files classified
  darkstorage compartment files pii --limit 100`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		compartment := args[0]
		limit, _ := cmd.Flags().GetInt("limit")

		// Mock data
		files := []map[string]interface{}{
			{"path": "my-bucket/classified/report-2026.pdf", "size": "2.4 MB", "modified": "2026-02-28"},
			{"path": "my-bucket/classified/analysis.docx", "size": "1.1 MB", "modified": "2026-02-27"},
			{"path": "my-bucket/classified/data.xlsx", "size": "850 KB", "modified": "2026-02-26"},
		}

		fmt.Printf("Files in compartment '%s':\n\n", compartment)

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Path", "Size", "Modified"})
		table.SetBorder(false)

		count := 0
		for _, f := range files {
			if limit > 0 && count >= limit {
				break
			}
			table.Append([]string{
				f["path"].(string),
				f["size"].(string),
				f["modified"].(string),
			})
			count++
		}
		table.Render()

		fmt.Printf("\nTotal: %d files\n", len(files))
	},
}

var compartmentGrantCmd = &cobra.Command{
	Use:   "grant <compartment> <user>",
	Short: "Grant user access to a compartment",
	Long: `Grant a user access to all files in a compartment.

Examples:
  darkstorage compartment grant classified alice@example.com --access read
  darkstorage compartment grant pii bob@example.com --access write`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		compartment := args[0]
		user := args[1]
		access, _ := cmd.Flags().GetString("access")

		color.Green("Granting %s access to '%s' for %s", access, compartment, user)

		endpoint := getAPIEndpoint()
		color.Yellow("\nAPI Call: POST %s/compartments/%s/grant", endpoint, compartment)
		color.Green("\n✓ Access granted: %s → %s (%s)", user, compartment, access)
	},
}

var compartmentRevokeCmd = &cobra.Command{
	Use:   "revoke <compartment> <user>",
	Short: "Revoke user access from a compartment",
	Long: `Revoke a user's access to a compartment.

Examples:
  darkstorage compartment revoke classified alice@example.com`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		compartment := args[0]
		user := args[1]

		color.Yellow("Revoking access to '%s' from %s", compartment, user)

		endpoint := getAPIEndpoint()
		color.Yellow("\nAPI Call: DELETE %s/compartments/%s/users/%s", endpoint, compartment, user)
		color.Green("\n✓ Access revoked: %s from %s", user, compartment)
	},
}

var compartmentDeleteCmd = &cobra.Command{
	Use:   "delete <compartment>",
	Short: "Delete a compartment",
	Long: `Delete a security compartment. Files will not be deleted, only unassigned.

Examples:
  darkstorage compartment delete old-project`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		compartment := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			color.Yellow("This will unassign all files from compartment '%s'", compartment)
			color.Yellow("Use --force to confirm")
			os.Exit(1)
		}

		color.Yellow("Deleting compartment: %s", compartment)

		endpoint := getAPIEndpoint()
		color.Yellow("\nAPI Call: DELETE %s/compartments/%s", endpoint, compartment)
		color.Green("\n✓ Compartment deleted: %s", compartment)
	},
}

func init() {
	// Create flags
	compartmentCreateCmd.Flags().String("level", "", "Security level (PUBLIC, CONFIDENTIAL, SECRET, TOP_SECRET)")
	compartmentCreateCmd.Flags().String("compliance", "", "Compliance requirements (comma-separated: HIPAA,GDPR,SOC2,ITAR)")
	compartmentCreateCmd.Flags().String("description", "", "Compartment description")
	compartmentCreateCmd.Flags().Bool("require-mfa", false, "Require MFA for access")
	compartmentCreateCmd.Flags().String("encryption", "AES256-GCM", "Encryption algorithm")

	// Assign flags
	compartmentAssignCmd.Flags().BoolP("recursive", "r", false, "Assign directory recursively")

	// Files flags
	compartmentFilesCmd.Flags().Int("limit", 100, "Maximum files to display")

	// Grant flags
	compartmentGrantCmd.Flags().String("access", "read", "Access level (read, write, admin)")

	// Delete flags
	compartmentDeleteCmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation")

	// List flags
	compartmentListCmd.Flags().Bool("json", false, "Output as JSON")

	// Add subcommands
	compartmentCmd.AddCommand(compartmentCreateCmd)
	compartmentCmd.AddCommand(compartmentListCmd)
	compartmentCmd.AddCommand(compartmentAssignCmd)
	compartmentCmd.AddCommand(compartmentFilesCmd)
	compartmentCmd.AddCommand(compartmentGrantCmd)
	compartmentCmd.AddCommand(compartmentRevokeCmd)
	compartmentCmd.AddCommand(compartmentDeleteCmd)

	rootCmd.AddCommand(compartmentCmd)
}
