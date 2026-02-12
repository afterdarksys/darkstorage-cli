package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var sharesCmd = &cobra.Command{
	Use:   "shares",
	Short: "Manage share links",
	Long:  `Create and manage share links for files and folders.`,
}

var shareCmd = &cobra.Command{
	Use:   "share [bucket/path]",
	Short: "Create a share link for a file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		expires, _ := cmd.Flags().GetString("expires")
		password, _ := cmd.Flags().GetString("password")
		maxDownloads, _ := cmd.Flags().GetInt("max-downloads")
		oneTime, _ := cmd.Flags().GetBool("one-time")

		client := newAPIClient()
		body := map[string]interface{}{
			"file_path": filePath,
		}

		if expires != "" {
			duration, _ := time.ParseDuration(expires)
			expiresAt := time.Now().Add(duration)
			body["expires_at"] = expiresAt.Format(time.RFC3339)
		}
		if password != "" {
			body["password"] = password
		}
		if maxDownloads > 0 {
			body["max_downloads"] = maxDownloads
		}
		if oneTime {
			body["one_time_use"] = true
		}

		resp, err := client.post("/sharing/links", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var result struct {
			URL   string `json:"url"`
			Token string `json:"token"`
		}
		json.Unmarshal(resp, &result)

		fmt.Printf("Share link created:\n%s\n", result.URL)
		if password != "" {
			fmt.Println("(Password protected)")
		}
	},
}

var sharesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all share links",
	Run: func(cmd *cobra.Command, args []string) {
		client := newAPIClient()
		resp, err := client.get("/sharing/links")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if viper.GetBool("json") {
			fmt.Println(string(resp))
			return
		}

		var result struct {
			Links []struct {
				ID            string     `json:"id"`
				FileName      string     `json:"file_name"`
				Token         string     `json:"token"`
				ExpiresAt     *time.Time `json:"expires_at"`
				DownloadCount int        `json:"download_count"`
				MaxDownloads  *int       `json:"max_downloads"`
				IsActive      bool       `json:"is_active"`
			} `json:"links"`
		}
		json.Unmarshal(resp, &result)

		if len(result.Links) == 0 {
			fmt.Println("No share links found")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tFILE\tDOWNLOADS\tEXPIRES\tSTATUS")
		for _, l := range result.Links {
			expires := "Never"
			if l.ExpiresAt != nil {
				expires = l.ExpiresAt.Format("2006-01-02")
			}
			downloads := fmt.Sprintf("%d", l.DownloadCount)
			if l.MaxDownloads != nil {
				downloads = fmt.Sprintf("%d/%d", l.DownloadCount, *l.MaxDownloads)
			}
			status := "Active"
			if !l.IsActive {
				status = "Inactive"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", l.ID[:8], l.FileName, downloads, expires, status)
		}
		w.Flush()
	},
}

var sharesRevokeCmd = &cobra.Command{
	Use:   "revoke [link-id]",
	Short: "Revoke a share link",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		linkID := args[0]
		client := newAPIClient()
		err := client.delete(fmt.Sprintf("/sharing/links/%s", linkID))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Share link revoked")
	},
}

var sharesStatsCmd = &cobra.Command{
	Use:   "stats [link-id]",
	Short: "Get statistics for a share link",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		linkID := args[0]
		client := newAPIClient()
		resp, err := client.get(fmt.Sprintf("/sharing/links/%s/stats", linkID))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if viper.GetBool("json") {
			fmt.Println(string(resp))
			return
		}

		var result struct {
			ViewCount      int   `json:"view_count"`
			DownloadCount  int   `json:"download_count"`
			BandwidthUsed  int64 `json:"bandwidth_used"`
			UniqueVisitors int   `json:"unique_visitors"`
		}
		json.Unmarshal(resp, &result)

		fmt.Printf("Views:           %d\n", result.ViewCount)
		fmt.Printf("Downloads:       %d\n", result.DownloadCount)
		fmt.Printf("Bandwidth Used:  %s\n", formatBytes(result.BandwidthUsed))
		fmt.Printf("Unique Visitors: %d\n", result.UniqueVisitors)
	},
}

func init() {
	rootCmd.AddCommand(shareCmd)
	rootCmd.AddCommand(sharesCmd)
	sharesCmd.AddCommand(sharesListCmd)
	sharesCmd.AddCommand(sharesRevokeCmd)
	sharesCmd.AddCommand(sharesStatsCmd)

	shareCmd.Flags().String("expires", "", "expiration time (e.g., 24h, 7d)")
	shareCmd.Flags().String("password", "", "password protect the link")
	shareCmd.Flags().Int("max-downloads", 0, "limit number of downloads")
	shareCmd.Flags().Bool("one-time", false, "one-time download link")
}
