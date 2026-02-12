package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var permsCmd = &cobra.Command{
	Use:   "perms",
	Short: "Manage file permissions",
	Long:  `Grant, revoke, and list file permissions for users and groups.`,
}

var permsListCmd = &cobra.Command{
	Use:   "list [file-id]",
	Short: "List permissions for a file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fileID := args[0]
		client := newAPIClient()
		resp, err := client.get(fmt.Sprintf("/files/%s/permissions", fileID))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if viper.GetBool("json") {
			fmt.Println(string(resp))
			return
		}

		var result struct {
			Permissions []struct {
				ID           string `json:"id"`
				GranteeType  string `json:"grantee_type"`
				GranteeName  string `json:"grantee_name"`
				GranteeEmail string `json:"grantee_email"`
				Level        string `json:"level"`
				CanRead      bool   `json:"can_read"`
				CanWrite     bool   `json:"can_write"`
				CanDelete    bool   `json:"can_delete"`
				CanShare     bool   `json:"can_share"`
			} `json:"permissions"`
		}
		json.Unmarshal(resp, &result)

		if len(result.Permissions) == 0 {
			fmt.Println("No permissions set")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TYPE\tGRANTEE\tLEVEL\tREAD\tWRITE\tDELETE\tSHARE")
		for _, p := range result.Permissions {
			grantee := p.GranteeName
			if p.GranteeEmail != "" {
				grantee = p.GranteeEmail
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%v\t%v\t%v\t%v\n",
				p.GranteeType, grantee, p.Level,
				boolMark(p.CanRead), boolMark(p.CanWrite),
				boolMark(p.CanDelete), boolMark(p.CanShare))
		}
		w.Flush()
	},
}

var permsGrantCmd = &cobra.Command{
	Use:   "grant [file-id]",
	Short: "Grant permission to a user or group",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fileID := args[0]
		user, _ := cmd.Flags().GetString("user")
		group, _ := cmd.Flags().GetString("group")
		level, _ := cmd.Flags().GetString("level")

		if user == "" && group == "" {
			fmt.Fprintln(os.Stderr, "Error: must specify --user or --group")
			os.Exit(1)
		}

		client := newAPIClient()
		body := map[string]interface{}{
			"level": level,
		}
		if user != "" {
			body["grantee_type"] = "user"
			body["grantee_email"] = user
		} else {
			body["grantee_type"] = "group"
			body["grantee_id"] = group
		}

		_, err := client.post(fmt.Sprintf("/files/%s/permissions", fileID), body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Permission granted")
	},
}

var permsRevokeCmd = &cobra.Command{
	Use:   "revoke [file-id] [permission-id]",
	Short: "Revoke a permission",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fileID := args[0]
		permID := args[1]

		client := newAPIClient()
		err := client.delete(fmt.Sprintf("/files/%s/permissions/%s", fileID, permID))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Permission revoked")
	},
}

var permsCheckCmd = &cobra.Command{
	Use:   "check [file-id]",
	Short: "Check your effective permissions on a file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fileID := args[0]
		client := newAPIClient()
		resp, err := client.get(fmt.Sprintf("/files/%s/permissions/effective", fileID))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var result struct {
			CanRead   bool `json:"can_read"`
			CanWrite  bool `json:"can_write"`
			CanDelete bool `json:"can_delete"`
			CanShare  bool `json:"can_share"`
			CanAdmin  bool `json:"can_admin"`
			IsOwner   bool `json:"is_owner"`
		}
		json.Unmarshal(resp, &result)

		fmt.Printf("Read:   %s\n", boolMark(result.CanRead))
		fmt.Printf("Write:  %s\n", boolMark(result.CanWrite))
		fmt.Printf("Delete: %s\n", boolMark(result.CanDelete))
		fmt.Printf("Share:  %s\n", boolMark(result.CanShare))
		fmt.Printf("Admin:  %s\n", boolMark(result.CanAdmin))
		fmt.Printf("Owner:  %s\n", boolMark(result.IsOwner))
	},
}

func boolMark(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func init() {
	rootCmd.AddCommand(permsCmd)
	permsCmd.AddCommand(permsListCmd)
	permsCmd.AddCommand(permsGrantCmd)
	permsCmd.AddCommand(permsRevokeCmd)
	permsCmd.AddCommand(permsCheckCmd)

	permsGrantCmd.Flags().String("user", "", "user email to grant permission to")
	permsGrantCmd.Flags().String("group", "", "group ID to grant permission to")
	permsGrantCmd.Flags().String("level", "viewer", "permission level: viewer, editor, admin")
}
