package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var trashCmd = &cobra.Command{
	Use:   "trash",
	Short: "Manage deleted files (SDMS)",
	Long:  `Secure Delete Management Service (SDMS) - recover deleted files within your tier's recovery window.`,
}

var trashListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recoverable deleted files",
	Run: func(cmd *cobra.Command, args []string) {
		client := newAPIClient()
		resp, err := client.get("/sdms/recoverable")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var result struct {
			Files []struct {
				ID             string    `json:"id"`
				OriginalName   string    `json:"original_name"`
				OriginalPath   string    `json:"original_path"`
				FileSize       int64     `json:"file_size"`
				DeletedAt      time.Time `json:"deleted_at"`
				ExpiresAt      time.Time `json:"expires_at"`
				HoursRemaining float64   `json:"hours_remaining"`
				Urgency        string    `json:"urgency"`
			} `json:"files"`
		}
		json.Unmarshal(resp, &result)

		if viper.GetBool("json") {
			fmt.Println(string(resp))
			return
		}

		if len(result.Files) == 0 {
			fmt.Println("No recoverable files found")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSIZE\tDELETED\tTIME LEFT\tURGENCY")
		for _, f := range result.Files {
			timeLeft := formatDuration(f.HoursRemaining)
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				f.ID[:8], f.OriginalName, formatBytes(f.FileSize),
				f.DeletedAt.Format("2006-01-02 15:04"), timeLeft, f.Urgency)
		}
		w.Flush()
	},
}

var trashRecoverCmd = &cobra.Command{
	Use:   "recover [file-id]",
	Short: "Recover a deleted file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fileID := args[0]
		toPath, _ := cmd.Flags().GetString("to")

		client := newAPIClient()
		body := map[string]string{}
		if toPath != "" {
			body["recover_to_path"] = toPath
		}

		resp, err := client.post(fmt.Sprintf("/sdms/%s/recover", fileID), body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var result struct {
			Message string `json:"message"`
		}
		json.Unmarshal(resp, &result)
		fmt.Println(result.Message)
	},
}

var trashPurgeCmd = &cobra.Command{
	Use:   "purge [file-id]",
	Short: "Permanently delete a file (no recovery)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fileID := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			fmt.Print("This will permanently delete the file. Continue? [y/N]: ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Aborted")
				return
			}
		}

		client := newAPIClient()
		err := client.delete(fmt.Sprintf("/sdms/%s", fileID))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("File permanently deleted")
	},
}

var trashInfoCmd = &cobra.Command{
	Use:   "info [file-id]",
	Short: "Show details about a deleted file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fileID := args[0]
		client := newAPIClient()
		resp, err := client.get(fmt.Sprintf("/sdms/%s", fileID))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if viper.GetBool("json") {
			fmt.Println(string(resp))
			return
		}

		var f struct {
			OriginalName   string    `json:"original_name"`
			OriginalPath   string    `json:"original_path"`
			FileSize       int64     `json:"file_size"`
			DeletedAt      time.Time `json:"deleted_at"`
			ExpiresAt      time.Time `json:"expires_at"`
			HoursRemaining float64   `json:"hours_remaining"`
			Status         string    `json:"status"`
		}
		json.Unmarshal(resp, &f)

		fmt.Printf("Name:           %s\n", f.OriginalName)
		fmt.Printf("Original Path:  %s\n", f.OriginalPath)
		fmt.Printf("Size:           %s\n", formatBytes(f.FileSize))
		fmt.Printf("Deleted:        %s\n", f.DeletedAt.Format(time.RFC3339))
		fmt.Printf("Expires:        %s\n", f.ExpiresAt.Format(time.RFC3339))
		fmt.Printf("Time Remaining: %s\n", formatDuration(f.HoursRemaining))
		fmt.Printf("Status:         %s\n", f.Status)
	},
}

func init() {
	rootCmd.AddCommand(trashCmd)
	trashCmd.AddCommand(trashListCmd)
	trashCmd.AddCommand(trashRecoverCmd)
	trashCmd.AddCommand(trashPurgeCmd)
	trashCmd.AddCommand(trashInfoCmd)

	trashRecoverCmd.Flags().String("to", "", "recover to specific path")
	trashPurgeCmd.Flags().Bool("force", false, "skip confirmation")
}

func formatDuration(hours float64) string {
	if hours <= 0 {
		return "Expired"
	}
	if hours < 1 {
		return fmt.Sprintf("%dm", int(hours*60))
	}
	if hours < 24 {
		return fmt.Sprintf("%dh", int(hours))
	}
	days := int(hours / 24)
	return fmt.Sprintf("%dd", days)
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

type apiClient struct {
	endpoint string
	apiKey   string
	http     *http.Client
}

func newAPIClient() *apiClient {
	return &apiClient{
		endpoint: viper.GetString("endpoint"),
		apiKey:   viper.GetString("api_key"),
		http:     &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *apiClient) get(path string) ([]byte, error) {
	req, _ := http.NewRequest("GET", c.endpoint+"/v1"+path, nil)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var body []byte
	resp.Body.Read(body)
	return body, nil
}

func (c *apiClient) post(path string, data interface{}) ([]byte, error) {
	// Implementation would use json.Marshal and http.Post
	return nil, nil
}

func (c *apiClient) delete(path string) error {
	req, _ := http.NewRequest("DELETE", c.endpoint+"/v1"+path, nil)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	_, err := c.http.Do(req)
	return err
}
