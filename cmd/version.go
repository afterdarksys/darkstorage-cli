package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version information (set by goreleaser)
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
	BuiltBy = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long: `Display version information for Dark Storage CLI.

Shows the current version, build commit, build date, and other metadata.`,
	Run: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")

		if verbose {
			fmt.Printf("Dark Storage CLI\n")
			fmt.Printf("  Version:    %s\n", Version)
			fmt.Printf("  Commit:     %s\n", Commit)
			fmt.Printf("  Built:      %s\n", Date)
			fmt.Printf("  Built by:   %s\n", BuiltBy)
			fmt.Printf("  Go version: %s\n", runtime.Version())
			fmt.Printf("  OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
		} else {
			fmt.Printf("darkstorage version %s\n", Version)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolP("verbose", "v", false, "Show detailed version information")
}
