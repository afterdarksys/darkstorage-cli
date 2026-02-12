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

func init() {
	rootCmd.AddCommand(auditCmd)
	auditCmd.AddCommand(auditListCmd)
	auditCmd.AddCommand(auditExportCmd)
	auditCmd.AddCommand(auditSummaryCmd)

	auditListCmd.Flags().String("type", "", "filter by event type")
	auditListCmd.Flags().String("from", "", "start date (YYYY-MM-DD)")
	auditListCmd.Flags().String("to", "", "end date (YYYY-MM-DD)")
	auditListCmd.Flags().Int("limit", 50, "number of results")

	auditExportCmd.Flags().String("format", "csv", "export format: csv, json")
	auditExportCmd.Flags().String("from", "", "start date (YYYY-MM-DD)")
	auditExportCmd.Flags().String("to", "", "end date (YYYY-MM-DD)")
}
