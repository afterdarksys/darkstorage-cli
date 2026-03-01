package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "View and export audit logs",
	Long:  `Access detailed activity logs for compliance and monitoring.`,
}

var auditListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent audit events",
	Run: func(cmd *cobra.Command, args []string) {
		eventType, _ := cmd.Flags().GetString("type")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		limit, _ := cmd.Flags().GetInt("limit")

		client := newAPIClient()
		path := fmt.Sprintf("/audit?limit=%d", limit)
		if eventType != "" {
			path += "&event_type=" + eventType
		}
		if from != "" {
			path += "&from=" + from
		}
		if to != "" {
			path += "&to=" + to
		}

		resp, err := client.get(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if viper.GetBool("json") {
			fmt.Println(string(resp))
			return
		}

		var result struct {
			Events []struct {
				ID           string    `json:"id"`
				EventType    string    `json:"event_type"`
				ResourceType string    `json:"resource_type"`
				ResourceName string    `json:"resource_name"`
				ActorEmail   string    `json:"actor_email"`
				IPAddress    string    `json:"ip_address"`
				CreatedAt    time.Time `json:"created_at"`
			} `json:"events"`
		}
		json.Unmarshal(resp, &result)

		if len(result.Events) == 0 {
			fmt.Println("No audit events found")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TIME\tEVENT\tRESOURCE\tUSER\tIP")
		for _, e := range result.Events {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				e.CreatedAt.Format("01-02 15:04"),
				e.EventType, e.ResourceName, e.ActorEmail, e.IPAddress)
		}
		w.Flush()
	},
}

var auditExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export audit logs",
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")

		client := newAPIClient()
		path := "/audit?limit=10000"
		if from != "" {
			path += "&from=" + from
		}
		if to != "" {
			path += "&to=" + to
		}

		resp, err := client.get(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var result struct {
			Events []struct {
				ID           string    `json:"id"`
				EventType    string    `json:"event_type"`
				ResourceType string    `json:"resource_type"`
				ResourceName string    `json:"resource_name"`
				ActorEmail   string    `json:"actor_email"`
				IPAddress    string    `json:"ip_address"`
				UserAgent    string    `json:"user_agent"`
				CreatedAt    time.Time `json:"created_at"`
			} `json:"events"`
		}
		json.Unmarshal(resp, &result)

		if format == "json" {
			fmt.Println(string(resp))
			return
		}

		// CSV output
		w := csv.NewWriter(os.Stdout)
		w.Write([]string{"timestamp", "event_type", "resource_type", "resource_name", "actor_email", "ip_address", "user_agent"})
		for _, e := range result.Events {
			w.Write([]string{
				e.CreatedAt.Format(time.RFC3339),
				e.EventType,
				e.ResourceType,
				e.ResourceName,
				e.ActorEmail,
				e.IPAddress,
				e.UserAgent,
			})
		}
		w.Flush()
	},
}

var auditSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Get audit summary statistics",
	Run: func(cmd *cobra.Command, args []string) {
		client := newAPIClient()
		resp, err := client.get("/audit/summary")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if viper.GetBool("json") {
			fmt.Println(string(resp))
			return
		}

		var result struct {
			TotalEvents    int `json:"total_events"`
			EventsByType   map[string]int `json:"events_by_type"`
			EventsToday    int `json:"events_today"`
			EventsThisWeek int `json:"events_this_week"`
		}
		json.Unmarshal(resp, &result)

		fmt.Printf("Total Events:     %d\n", result.TotalEvents)
		fmt.Printf("Events Today:     %d\n", result.EventsToday)
		fmt.Printf("Events This Week: %d\n", result.EventsThisWeek)
		if len(result.EventsByType) > 0 {
			fmt.Println("\nEvents by Type:")
			for k, v := range result.EventsByType {
				fmt.Printf("  %s: %d\n", k, v)
			}
		}
	},
}

var auditFileCmd = &cobra.Command{
	Use:   "file <path>",
	Short: "Show complete audit history for a file",
	Long: `Show all actions performed on a specific file.

Examples:
  darkstorage audit file my-bucket/contract.pdf
  darkstorage audit file my-bucket/contract.pdf --accessed-by
  darkstorage audit file my-bucket/contract.pdf --downloads`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Implementation would query backend
		fmt.Printf("Audit history for: %s\n", args[0])
		fmt.Println("(This would show complete file history from backend)")
	},
}

var auditUserCmd = &cobra.Command{
	Use:   "user <email>",
	Short: "Show all activity by a user",
	Long: `Show all actions performed by a specific user.

Examples:
  darkstorage audit user alice@example.com
  darkstorage audit user alice@example.com --files
  darkstorage audit user alice@example.com --logins`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Implementation would query backend
		fmt.Printf("Activity for user: %s\n", args[0])
		fmt.Println("(This would show user activity from backend)")
	},
}

var auditStreamCmd = &cobra.Command{
	Use:   "stream",
	Short: "Stream audit events in real-time",
	Long: `Watch audit events as they happen (live tail).

Examples:
  darkstorage audit stream
  darkstorage audit stream --type FILE_DOWNLOAD,FILE_DELETE
  darkstorage audit stream --category security`,
	Run: func(cmd *cobra.Command, args []string) {
		// Implementation would use WebSocket/SSE
		fmt.Println("Streaming audit events... (Ctrl+C to stop)")
		fmt.Println("(This would connect to real-time event stream)")
	},
}

var auditViolationsCmd = &cobra.Command{
	Use:   "violations",
	Short: "Show compliance violations",
	Long: `List all detected compliance violations.

Examples:
  darkstorage audit violations
  darkstorage audit violations --severity high`,
	Run: func(cmd *cobra.Command, args []string) {
		// Implementation would query backend
		fmt.Println("Compliance Violations:")
		fmt.Println("(This would show violations from backend)")
	},
}

func init() {
	rootCmd.AddCommand(auditCmd)
	auditCmd.AddCommand(auditListCmd)
	auditCmd.AddCommand(auditExportCmd)
	auditCmd.AddCommand(auditSummaryCmd)
	auditCmd.AddCommand(auditFileCmd)
	auditCmd.AddCommand(auditUserCmd)
	auditCmd.AddCommand(auditStreamCmd)
	auditCmd.AddCommand(auditViolationsCmd)

	// List command flags
	auditListCmd.Flags().String("type", "", "Filter by event type (FILE_DOWNLOAD, PERMISSION_GRANT, etc)")
	auditListCmd.Flags().String("category", "", "Filter by category (file, permission, auth, security, admin)")
	auditListCmd.Flags().String("user", "", "Filter by user email")
	auditListCmd.Flags().String("resource", "", "Filter by resource path")
	auditListCmd.Flags().String("ip", "", "Filter by IP address")
	auditListCmd.Flags().String("country", "", "Filter by country code (US, CN, etc)")
	auditListCmd.Flags().String("compartment", "", "Filter by compartment name")
	auditListCmd.Flags().String("from", "", "Start date (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)")
	auditListCmd.Flags().String("to", "", "End date (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)")
	auditListCmd.Flags().String("since", "", "Relative time (24h, 7d, 30d)")
	auditListCmd.Flags().String("before", "", "Before timestamp")
	auditListCmd.Flags().Int("limit", 50, "Number of results")
	auditListCmd.Flags().Bool("success-only", false, "Show only successful events")
	auditListCmd.Flags().Bool("failed-only", false, "Show only failed events")
	auditListCmd.Flags().Int("risk-score", 0, "Minimum risk score (0-100)")

	// Export command flags
	auditExportCmd.Flags().String("format", "csv", "Export format: csv, json, pdf")
	auditExportCmd.Flags().String("from", "", "Start date (YYYY-MM-DD)")
	auditExportCmd.Flags().String("to", "", "End date (YYYY-MM-DD)")
	auditExportCmd.Flags().String("output", "", "Output file path")
	auditExportCmd.Flags().String("type", "", "Filter by event type")
	auditExportCmd.Flags().String("user", "", "Filter by user")

	// File command flags
	auditFileCmd.Flags().Bool("accessed-by", false, "Show who accessed the file")
	auditFileCmd.Flags().Bool("downloads", false, "Show download history")
	auditFileCmd.Flags().Bool("permission-changes", false, "Show permission changes")

	// User command flags
	auditUserCmd.Flags().Bool("files", false, "Show files accessed")
	auditUserCmd.Flags().Bool("logins", false, "Show login history")
	auditUserCmd.Flags().Bool("permissions-granted", false, "Show permissions granted by user")

	// Stream command flags
	auditStreamCmd.Flags().String("type", "", "Filter event types")
	auditStreamCmd.Flags().String("category", "", "Filter by category")

	// Violations command flags
	auditViolationsCmd.Flags().String("severity", "", "Filter by severity (low, medium, high, critical)")
	auditViolationsCmd.Flags().String("type", "", "Filter by violation type")
}
