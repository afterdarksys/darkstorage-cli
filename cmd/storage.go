package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// ls command
var lsCmd = &cobra.Command{
	Use:   "ls [bucket/path]",
	Short: "List files and buckets",
	Long: `List files in a bucket or list all buckets.

Examples:
  darkstorage ls                    # List all buckets
  darkstorage ls my-bucket          # List files in bucket
  darkstorage ls my-bucket/folder/  # List files in folder`,
	Run: func(cmd *cobra.Command, args []string) {
		recursive, _ := cmd.Flags().GetBool("recursive")
		long, _ := cmd.Flags().GetBool("long")

		if len(args) == 0 {
			// List buckets
			fmt.Println("Buckets:")
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Objects", "Size", "Created"})
			table.SetBorder(false)
			table.Append([]string{"my-bucket", "1,234", "2.4 GB", "2024-01-01"})
			table.Append([]string{"backups", "45", "12.8 GB", "2023-12-15"})
			table.Append([]string{"assets", "892", "890 MB", "2023-11-20"})
			table.Render()
			return
		}

		path := args[0]
		bucket := strings.Split(path, "/")[0]

		if long {
			fmt.Printf("Contents of %s:\n", path)
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Size", "Modified", "Type"})
			table.SetBorder(false)
			table.Append([]string{"backups/", "-", "2024-01-05", "folder"})
			table.Append([]string{"assets/", "-", "2024-01-04", "folder"})
			table.Append([]string{"config.json", "4 KB", "2024-01-03", "file"})
			table.Append([]string{"readme.md", "2 KB", "2024-01-02", "file"})
			table.Render()
		} else {
			fmt.Printf("backups/  assets/  config.json  readme.md\n")
		}

		_ = bucket
		_ = recursive
	},
}

// put command
var putCmd = &cobra.Command{
	Use:   "put <source> <destination>",
	Short: "Upload files to storage",
	Long: `Upload files or directories to Dark Storage.

Examples:
  darkstorage put ./file.txt my-bucket/
  darkstorage put ./folder/ my-bucket/folder/ --recursive
  darkstorage put ./file.txt my-bucket/custom-name.txt`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		dest := args[1]
		recursive, _ := cmd.Flags().GetBool("recursive")

		info, err := os.Stat(source)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		if info.IsDir() && !recursive {
			color.Yellow("Warning: %s is a directory. Use --recursive to upload directories.", source)
			os.Exit(1)
		}

		// Simulate upload
		fmt.Printf("Uploading %s to %s...\n", source, dest)
		color.Green("✓ Upload complete: %s (%s)", filepath.Base(source), humanize.Bytes(uint64(info.Size())))
	},
}

// get command
var getCmd = &cobra.Command{
	Use:   "get <source> [destination]",
	Short: "Download files from storage",
	Long: `Download files from Dark Storage.

Examples:
  darkstorage get my-bucket/file.txt ./
  darkstorage get my-bucket/folder/ ./ --recursive`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		dest := "."
		if len(args) > 1 {
			dest = args[1]
		}

		fmt.Printf("Downloading %s to %s...\n", source, dest)
		color.Green("✓ Download complete: %s", filepath.Base(source))
	},
}

// rm command
var rmCmd = &cobra.Command{
	Use:   "rm <path>",
	Short: "Remove files or buckets",
	Long: `Remove files or empty buckets.

Examples:
  darkstorage rm my-bucket/file.txt
  darkstorage rm my-bucket/folder/ --recursive
  darkstorage rm my-bucket --force  # Delete bucket and contents`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		force, _ := cmd.Flags().GetBool("force")
		recursive, _ := cmd.Flags().GetBool("recursive")

		if force {
			fmt.Printf("Deleting %s and all contents...\n", path)
		} else if recursive {
			fmt.Printf("Deleting %s recursively...\n", path)
		} else {
			fmt.Printf("Deleting %s...\n", path)
		}

		color.Green("✓ Deleted: %s", path)
	},
}

// cat command
var catCmd = &cobra.Command{
	Use:   "cat <path>",
	Short: "Display file contents",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		fmt.Printf("Contents of %s:\n", path)
		fmt.Println("{ \"example\": \"file content\" }")
	},
}

// cp command
var cpCmd = &cobra.Command{
	Use:   "cp <source> <destination>",
	Short: "Copy files between locations",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		dest := args[1]
		fmt.Printf("Copying %s to %s...\n", source, dest)
		color.Green("✓ Copied successfully")
	},
}

// mv command
var mvCmd = &cobra.Command{
	Use:   "mv <source> <destination>",
	Short: "Move/rename files",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		dest := args[1]
		fmt.Printf("Moving %s to %s...\n", source, dest)
		color.Green("✓ Moved successfully")
	},
}

// search command
var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for files",
	Long: `Search for files by name, metadata, or content type.

Examples:
  darkstorage search "backup"
  darkstorage search --type image/png
  darkstorage search --metadata key=value`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")
		fmt.Printf("Searching for: %s\n\n", query)

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"File", "Bucket", "Size", "Modified"})
		table.SetBorder(false)
		table.Append([]string{"backup-2024.tar.gz", "backups", "1.2 GB", "2024-01-05"})
		table.Append([]string{"backup-2023.tar.gz", "backups", "980 MB", "2023-12-31"})
		table.Render()
	},
}

// hashsearch command
var hashSearchCmd = &cobra.Command{
	Use:   "hashsearch <hash>",
	Short: "Search files by hash",
	Long: `Search for files by their SHA-256 hash.

Examples:
  darkstorage hashsearch abc123def456...`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hash := args[0]
		fmt.Printf("Searching for hash: %s\n\n", hash[:16]+"...")

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"File", "Bucket", "Size"})
		table.SetBorder(false)
		table.Append([]string{"document.pdf", "my-bucket", "2.4 MB"})
		table.Render()
	},
}

// metadata command
var metadataCmd = &cobra.Command{
	Use:   "metadata <path>",
	Short: "View or set file metadata",
	Long: `View or modify file metadata.

Examples:
  darkstorage metadata my-bucket/file.txt
  darkstorage metadata my-bucket/file.txt --set key=value`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		set, _ := cmd.Flags().GetStringArray("set")

		if len(set) > 0 {
			for _, kv := range set {
				fmt.Printf("Setting metadata: %s\n", kv)
			}
			color.Green("✓ Metadata updated")
			return
		}

		fmt.Printf("Metadata for %s:\n", path)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Key", "Value"})
		table.SetBorder(false)
		table.Append([]string{"Content-Type", "text/plain"})
		table.Append([]string{"Size", "4 KB"})
		table.Append([]string{"ETag", "abc123def456"})
		table.Append([]string{"Last-Modified", "2024-01-05T10:30:00Z"})
		table.Append([]string{"custom-tag", "important"})
		table.Render()
	},
}

// attrs command
var attrsCmd = &cobra.Command{
	Use:   "attrs <path>",
	Short: "View or set file attributes",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		fmt.Printf("Attributes for %s:\n", path)
		fmt.Println("  read-only: false")
		fmt.Println("  hidden: false")
		fmt.Println("  immutable: false")
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(putCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(catCmd)
	rootCmd.AddCommand(cpCmd)
	rootCmd.AddCommand(mvCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(hashSearchCmd)
	rootCmd.AddCommand(metadataCmd)
	rootCmd.AddCommand(attrsCmd)

	// ls flags
	lsCmd.Flags().BoolP("recursive", "r", false, "list recursively")
	lsCmd.Flags().BoolP("long", "l", false, "long listing format")

	// put flags
	putCmd.Flags().BoolP("recursive", "r", false, "upload directories recursively")
	putCmd.Flags().String("content-type", "", "set content type")

	// get flags
	getCmd.Flags().BoolP("recursive", "r", false, "download directories recursively")

	// rm flags
	rmCmd.Flags().BoolP("recursive", "r", false, "delete recursively")
	rmCmd.Flags().BoolP("force", "f", false, "force delete (buckets)")

	// metadata flags
	metadataCmd.Flags().StringArray("set", nil, "set metadata key=value")

	// search flags
	searchCmd.Flags().String("type", "", "filter by content type")
	searchCmd.Flags().StringArray("metadata", nil, "filter by metadata key=value")
}
