package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var apiCmd = &cobra.Command{
	Use:   "api <method> <path> [body]",
	Short: "Make direct API calls",
	Long: `Make direct HTTP requests to the DarkStorage API.

The command automatically handles authentication using your stored credentials.

Examples:
  # GET request
  darkstorage api GET /buckets

  # POST request with JSON body
  darkstorage api POST /buckets '{"name":"my-bucket"}'

  # PUT request with file body
  darkstorage api PUT /objects/my-bucket/file.txt @./local-file.txt

  # DELETE request
  darkstorage api DELETE /buckets/my-bucket

  # Custom headers
  darkstorage api GET /files -H "X-Custom: value"`,
	Args: cobra.RangeArgs(2, 3),
	Run: func(cmd *cobra.Command, args []string) {
		method := strings.ToUpper(args[0])
		path := args[1]
		var body string
		if len(args) > 2 {
			body = args[2]
		}

		headers, _ := cmd.Flags().GetStringSlice("header")
		pretty, _ := cmd.Flags().GetBool("pretty")
		verbose, _ := cmd.Flags().GetBool("verbose")
		includeHeaders, _ := cmd.Flags().GetBool("include")

		if err := makeAPIRequest(method, path, body, headers, pretty, verbose, includeHeaders); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}
	},
}

func makeAPIRequest(method, path, body string, headers []string, pretty, verbose, includeHeaders bool) error {
	// Get API endpoint from config
	endpoint := getAPIEndpoint()

	// Build full URL
	url := endpoint + path
	if verbose {
		color.Cyan("→ %s %s", method, url)
	}

	// Prepare request body
	var bodyReader io.Reader
	if body != "" {
		// Check if body is a file reference (@filename)
		if strings.HasPrefix(body, "@") {
			filename := strings.TrimPrefix(body, "@")
			file, err := os.Open(filename)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", filename, err)
			}
			defer file.Close()
			bodyReader = file

			if verbose {
				stat, _ := file.Stat()
				color.Cyan("→ Body: %s (%d bytes)", filename, stat.Size())
			}
		} else {
			bodyReader = strings.NewReader(body)
			if verbose {
				color.Cyan("→ Body: %s", body)
			}
		}
	}

	// Create request
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication
	token := getAuthToken()
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
		if verbose {
			color.Cyan("→ Authorization: Bearer %s...", token[:min(20, len(token))])
		}
	}

	// Add custom headers
	for _, header := range headers {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			req.Header.Set(key, value)
			if verbose {
				color.Cyan("→ %s: %s", key, value)
			}
		}
	}

	// Set default content type for JSON bodies
	if body != "" && !strings.HasPrefix(body, "@") && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Print response headers if requested
	if includeHeaders {
		color.Yellow("\n← HTTP/%d.%d %s", resp.ProtoMajor, resp.ProtoMinor, resp.Status)
		for key, values := range resp.Header {
			for _, value := range values {
				color.Yellow("← %s: %s", key, value)
			}
		}
		fmt.Println()
	} else if verbose {
		color.Yellow("← %s", resp.Status)
	}

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Print response
	if len(respBody) > 0 {
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/json") && pretty {
			// Pretty print JSON
			var jsonData interface{}
			if err := json.Unmarshal(respBody, &jsonData); err == nil {
				prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
				if err == nil {
					fmt.Println(string(prettyJSON))
				} else {
					fmt.Println(string(respBody))
				}
			} else {
				fmt.Println(string(respBody))
			}
		} else {
			fmt.Println(string(respBody))
		}
	}

	// Exit with error code if request failed
	if resp.StatusCode >= 400 {
		os.Exit(1)
	}

	return nil
}

func getAPIEndpoint() string {
	// Try to get from flag
	if endpoint := rootCmd.Flag("endpoint").Value.String(); endpoint != "" {
		return endpoint
	}
	// Default
	return "https://api.darkstorage.io"
}

func getAuthToken() string {
	// Try to get from API key flag
	if apiKey := rootCmd.Flag("api-key").Value.String(); apiKey != "" {
		return apiKey
	}

	// Try to load from config
	cfg, err := loadConfig()
	if err == nil && cfg.AccessToken != "" {
		return cfg.AccessToken
	}

	return ""
}

func loadConfig() (*Config, error) {
	configPath := getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func getConfigPath() string {
	if configFile := rootCmd.Flag("config").Value.String(); configFile != "" {
		return configFile
	}

	home, _ := os.UserHomeDir()
	return home + "/.darkstorage/config.json"
}

type Config struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Endpoint     string `json:"endpoint"`
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func init() {
	apiCmd.Flags().StringSliceP("header", "H", []string{}, "Add custom header (can be used multiple times)")
	apiCmd.Flags().BoolP("pretty", "p", true, "Pretty print JSON responses")
	apiCmd.Flags().BoolP("include", "i", false, "Include response headers")
	apiCmd.Flags().BoolP("verbose", "v", false, "Verbose output (show request details)")

	rootCmd.AddCommand(apiCmd)
}
