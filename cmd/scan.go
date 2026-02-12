package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Security scanning commands",
	Long:  `Scan files for malware, viruses, and sensitive data. Manage quarantine.`,
}

var scanFileCmd = &cobra.Command{
	Use:   "file [bucket/path]",
	Short: "Scan a file for threats",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		scanType, _ := cmd.Flags().GetString("type")

		client := newAPIClient()
		body := map[string]interface{}{
			"scan_type": scanType,
			"priority":  5,
		}

		resp, err := client.post(fmt.Sprintf("/files/%s/scan", filePath), body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var result struct {
			Message string `json:"message"`
			QueueID string `json:"queue_id"`
		}
		json.Unmarshal(resp, &result)
		fmt.Println(result.Message)
		if result.QueueID != "" {
			fmt.Printf("Queue ID: %s\n", result.QueueID)
		}
	},
}

var scanStatusCmd = &cobra.Command{
	Use:   "status [file-id]",
	Short: "Get scan status for a file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fileID := args[0]
		client := newAPIClient()
		resp, err := client.get(fmt.Sprintf("/files/%s/scan-result", fileID))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if viper.GetBool("json") {
			fmt.Println(string(resp))
			return
		}

		var result struct {
			Status     string `json:"status"`
			ScanEngine string `json:"scan_engine"`
			Severity   string `json:"severity"`
			Threats    []struct {
				Name     string `json:"name"`
				Category string `json:"category"`
			} `json:"threats"`
		}
		json.Unmarshal(resp, &result)

		fmt.Printf("Status:   %s\n", result.Status)
		fmt.Printf("Engine:   %s\n", result.ScanEngine)
		fmt.Printf("Severity: %s\n", result.Severity)
		if len(result.Threats) > 0 {
			fmt.Println("Threats:")
			for _, t := range result.Threats {
				fmt.Printf("  - %s (%s)\n", t.Name, t.Category)
			}
		}
	},
}

var scanThreatsCmd = &cobra.Command{
	Use:   "threats",
	Short: "List all detected threats",
	Run: func(cmd *cobra.Command, args []string) {
		client := newAPIClient()
		resp, err := client.get("/scanner/threats")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if viper.GetBool("json") {
			fmt.Println(string(resp))
			return
		}

		var result struct {
			Threats []struct {
				FileID   string `json:"file_id"`
				FileName string `json:"file_name"`
				Status   string `json:"status"`
				Severity string `json:"severity"`
			} `json:"threats"`
		}
		json.Unmarshal(resp, &result)

		if len(result.Threats) == 0 {
			fmt.Println("No threats detected")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "FILE\tSTATUS\tSEVERITY")
		for _, t := range result.Threats {
			fmt.Fprintf(w, "%s\t%s\t%s\n", t.FileName, t.Status, t.Severity)
		}
		w.Flush()
	},
}

var scanQuarantineCmd = &cobra.Command{
	Use:   "quarantine",
	Short: "Manage quarantined files",
}

var scanQuarantineListCmd = &cobra.Command{
	Use:   "list",
	Short: "List quarantined files",
	Run: func(cmd *cobra.Command, args []string) {
		client := newAPIClient()
		resp, err := client.get("/scanner/quarantine")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if viper.GetBool("json") {
			fmt.Println(string(resp))
			return
		}

		var result struct {
			Files []struct {
				ID               string `json:"id"`
				OriginalName     string `json:"original_name"`
				QuarantineReason string `json:"quarantine_reason"`
				FileSize         int64  `json:"file_size"`
			} `json:"files"`
		}
		json.Unmarshal(resp, &result)

		if len(result.Files) == 0 {
			fmt.Println("No quarantined files")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tREASON\tSIZE")
		for _, f := range result.Files {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", f.ID[:8], f.OriginalName, f.QuarantineReason, formatBytes(f.FileSize))
		}
		w.Flush()
	},
}

var scanQuarantineReleaseCmd = &cobra.Command{
	Use:   "release [quarantine-id]",
	Short: "Release a file from quarantine",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		qID := args[0]
		client := newAPIClient()
		_, err := client.post(fmt.Sprintf("/scanner/quarantine/%s/release", qID), nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("File released from quarantine")
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.AddCommand(scanFileCmd)
	scanCmd.AddCommand(scanStatusCmd)
	scanCmd.AddCommand(scanThreatsCmd)
	scanCmd.AddCommand(scanQuarantineCmd)
	scanQuarantineCmd.AddCommand(scanQuarantineListCmd)
	scanQuarantineCmd.AddCommand(scanQuarantineReleaseCmd)

	scanFileCmd.Flags().String("type", "antivirus", "scan type: antivirus, magic, dlp")
}
