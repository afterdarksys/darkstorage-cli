package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var groupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "Manage groups and team access",
	Long:  `Create and manage groups for team collaboration and shared file access.`,
}

var groupsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List your groups",
	Run: func(cmd *cobra.Command, args []string) {
		client := newAPIClient()
		resp, err := client.get("/groups")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if viper.GetBool("json") {
			fmt.Println(string(resp))
			return
		}

		var result struct {
			Groups []struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				MemberCount int    `json:"member_count"`
				MyRole      string `json:"my_role"`
			} `json:"groups"`
		}
		json.Unmarshal(resp, &result)

		if len(result.Groups) == 0 {
			fmt.Println("No groups found")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tMEMBERS\tMY ROLE")
		for _, g := range result.Groups {
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", g.ID[:8], g.Name, g.MemberCount, g.MyRole)
		}
		w.Flush()
	},
}

var groupsCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new group",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		description, _ := cmd.Flags().GetString("description")

		client := newAPIClient()
		body := map[string]string{
			"name":        name,
			"description": description,
		}

		resp, err := client.post("/groups", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var result struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		json.Unmarshal(resp, &result)
		fmt.Printf("Created group '%s' (ID: %s)\n", result.Name, result.ID[:8])
	},
}

var groupsAddMemberCmd = &cobra.Command{
	Use:   "add-member [group-id] [email]",
	Short: "Add a member to a group",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		groupID := args[0]
		email := args[1]
		role, _ := cmd.Flags().GetString("role")

		client := newAPIClient()
		body := map[string]string{
			"email": email,
			"role":  role,
		}

		_, err := client.post(fmt.Sprintf("/groups/%s/members", groupID), body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Invitation sent to %s\n", email)
	},
}

var groupsRemoveMemberCmd = &cobra.Command{
	Use:   "remove-member [group-id] [email]",
	Short: "Remove a member from a group",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		groupID := args[0]
		email := args[1]

		client := newAPIClient()
		err := client.delete(fmt.Sprintf("/groups/%s/members/%s", groupID, email))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Member removed from group")
	},
}

var groupsMembersCmd = &cobra.Command{
	Use:   "members [group-id]",
	Short: "List group members",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupID := args[0]
		client := newAPIClient()
		resp, err := client.get(fmt.Sprintf("/groups/%s/members", groupID))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if viper.GetBool("json") {
			fmt.Println(string(resp))
			return
		}

		var result struct {
			Members []struct {
				Email  string `json:"email"`
				Name   string `json:"name"`
				Role   string `json:"role"`
				Status string `json:"status"`
			} `json:"members"`
		}
		json.Unmarshal(resp, &result)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "EMAIL\tNAME\tROLE\tSTATUS")
		for _, m := range result.Members {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", m.Email, m.Name, m.Role, m.Status)
		}
		w.Flush()
	},
}

var groupsDeleteCmd = &cobra.Command{
	Use:   "delete [group-id]",
	Short: "Delete a group",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupID := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			fmt.Print("Delete this group? [y/N]: ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Aborted")
				return
			}
		}

		client := newAPIClient()
		err := client.delete(fmt.Sprintf("/groups/%s", groupID))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Group deleted")
	},
}

func init() {
	rootCmd.AddCommand(groupsCmd)
	groupsCmd.AddCommand(groupsListCmd)
	groupsCmd.AddCommand(groupsCreateCmd)
	groupsCmd.AddCommand(groupsAddMemberCmd)
	groupsCmd.AddCommand(groupsRemoveMemberCmd)
	groupsCmd.AddCommand(groupsMembersCmd)
	groupsCmd.AddCommand(groupsDeleteCmd)

	groupsCreateCmd.Flags().String("description", "", "group description")
	groupsAddMemberCmd.Flags().String("role", "member", "member role: owner, admin, member")
	groupsDeleteCmd.Flags().Bool("force", false, "skip confirmation")
}
