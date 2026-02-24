package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/darkstorage/cli/internal/config"
	"github.com/darkstorage/cli/internal/storage"
	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var storageBackend storage.StorageBackend

// initStorage initializes the storage backend
func initStorage() error {
	if storageBackend != nil {
		return nil
	}

	cfg, err := config.LoadStorageConfig()
	if err != nil {
		return fmt.Errorf("failed to load storage config: %w", err)
	}

	backend, err := storage.NewTraditionalBackend(&storage.TraditionalConfig{
		Endpoint:  cfg.Endpoint,
		AccessKey: cfg.AccessKey,
		SecretKey: cfg.SecretKey,
		UseSSL:    cfg.UseSSL,
		Region:    cfg.Region,
	})
	if err != nil {
		return fmt.Errorf("failed to create storage backend: %w", err)
	}

	storageBackend = backend
	return nil
}

// ls command
var lsCmd = &cobra.Command{
	Use:   "ls [bucket/path]",
	Short: "List files and buckets",
	Long: `List files in a bucket or list all buckets.

Examples:
  darkstorage ls                    # List all buckets
  darkstorage ls test-bucket          # List files in bucket
  darkstorage ls test-bucket/folder/  # List files in folder`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		recursive, _ := cmd.Flags().GetBool("recursive")
		long, _ := cmd.Flags().GetBool("long")

		if len(args) == 0 {
			// List buckets
			buckets, err := storageBackend.ListBuckets(ctx)
			if err != nil {
				color.Red("Error listing buckets: %v", err)
				os.Exit(1)
			}

			if len(buckets) == 0 {
				fmt.Println("No buckets found.")
				fmt.Println("\nCreate a bucket with: darkstorage mb <bucket-name>")
				return
			}

			fmt.Println("Buckets:")
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Created"})
			table.SetBorder(false)

			for _, bucket := range buckets {
				table.Append([]string{
					bucket.Name,
					bucket.CreatedAt.Format("2006-01-02"),
				})
			}
			table.Render()
			return
		}

		// List files in bucket/path
		path := args[0]

		opts := &storage.ListOptions{
			Recursive: recursive,
		}

		files, err := storageBackend.List(ctx, path, opts)
		if err != nil {
			color.Red("Error listing files: %v", err)
			os.Exit(1)
		}

		if len(files) == 0 {
			fmt.Printf("No files found in %s\n", path)
			return
		}

		if long {
			fmt.Printf("Contents of %s:\n", path)
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Size", "Modified", "Type"})
			table.SetBorder(false)

			for _, file := range files {
				fileType := "file"
				size := humanize.Bytes(uint64(file.Size))
				if file.IsDir {
					fileType = "folder"
					size = "-"
				}

				table.Append([]string{
					file.Name,
					size,
					file.ModifiedAt.Format("2006-01-02 15:04"),
					fileType,
				})
			}
			table.Render()
		} else {
			for _, file := range files {
				if file.IsDir {
					fmt.Printf("%s/  ", file.Name)
				} else {
					fmt.Printf("%s  ", file.Name)
				}
			}
			fmt.Println()
		}
	},
}

// mb command (make bucket)
var mbCmd = &cobra.Command{
	Use:   "mb <bucket-name>",
	Short: "Create a new bucket",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		bucketName := args[0]

		if err := storageBackend.CreateBucket(ctx, bucketName); err != nil {
			color.Red("Error creating bucket: %v", err)
			os.Exit(1)
		}

		color.Green("✓ Bucket created: %s", bucketName)
	},
}

// put command
var putCmd = &cobra.Command{
	Use:   "put <source> <destination>",
	Short: "Upload files to storage",
	Long: `Upload files or directories to Dark Storage.

Examples:
  darkstorage put ./file.txt test-bucket/
  darkstorage put ./folder/ test-bucket/folder/ --recursive
  darkstorage put ./file.txt test-bucket/custom-name.txt`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

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

		ctx := context.Background()

		if info.IsDir() {
			// Recursive directory upload
			uploadDir(ctx, source, dest, storageBackend)
			return
		}

		// Upload single file
		uploadFile(ctx, source, dest, info.Size(), storageBackend)
	},
}

// get command
var getCmd = &cobra.Command{
	Use:   "get <source> [destination]",
	Short: "Download files from storage",
	Long: `Download files from Dark Storage.

Examples:
  darkstorage get test-bucket/file.txt ./
  darkstorage get test-bucket/folder/ ./ --recursive`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		source := args[0]
		dest := "."
		if len(args) > 1 {
			dest = args[1]
		}

		recursive, _ := cmd.Flags().GetBool("recursive")
		ctx := context.Background()

		// Check if recursive download is requested
		if recursive || strings.HasSuffix(source, "/") {
			// Recursive directory download
			downloadDir(ctx, source, dest, storageBackend)
			return
		}

		// Get file info first
		stat, err := storageBackend.Stat(ctx, source)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		// Download single file
		downloadFile(ctx, source, dest, stat.Size, storageBackend)
	},
}

// rm command
var rmCmd = &cobra.Command{
	Use:   "rm <path>",
	Short: "Remove files or buckets",
	Long: `Remove files or empty buckets.

Examples:
  darkstorage rm test-bucket/file.txt
  darkstorage rm test-bucket/folder/ --recursive
  darkstorage rm test-bucket --force  # Delete bucket`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		path := args[0]
		force, _ := cmd.Flags().GetBool("force")
		recursive, _ := cmd.Flags().GetBool("recursive")

		ctx := context.Background()

		// Check if it's a bucket (no /)
		if !strings.Contains(path, "/") {
			// Delete bucket
			if !force {
				color.Yellow("Use --force to delete bucket: %s", path)
				os.Exit(1)
			}

			if err := storageBackend.DeleteBucket(ctx, path); err != nil {
				color.Red("Error deleting bucket: %v", err)
				os.Exit(1)
			}

			color.Green("✓ Bucket deleted: %s", path)
			return
		}

		// Delete file(s)
		if recursive || strings.HasSuffix(path, "/") {
			// Recursive delete
			deleteDir(ctx, path, storageBackend)
			return
		}

		// Delete single file
		if err := storageBackend.Delete(ctx, path); err != nil {
			color.Red("Error deleting file: %v", err)
			os.Exit(1)
		}

		color.Green("✓ Deleted: %s", path)
	},
}

// cp command
var cpCmd = &cobra.Command{
	Use:   "cp <source> <destination>",
	Short: "Copy files between locations",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		source := args[0]
		dest := args[1]

		ctx := context.Background()

		if err := storageBackend.Copy(ctx, source, dest); err != nil {
			color.Red("Error copying file: %v", err)
			os.Exit(1)
		}

		color.Green("✓ Copied: %s → %s", source, dest)
	},
}

// mv command
var mvCmd = &cobra.Command{
	Use:   "mv <source> <destination>",
	Short: "Move/rename files",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		source := args[0]
		dest := args[1]

		ctx := context.Background()

		if err := storageBackend.Move(ctx, source, dest); err != nil {
			color.Red("Error moving file: %v", err)
			os.Exit(1)
		}

		color.Green("✓ Moved: %s → %s", source, dest)
	},
}

// cat command
var catCmd = &cobra.Command{
	Use:   "cat <path>",
	Short: "Display file contents",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		path := args[0]
		ctx := context.Background()

		// Download to stdout
		result, err := storageBackend.Download(ctx, path, os.Stdout, nil)
		if err != nil {
			color.Red("Error reading file: %v", err)
			os.Exit(1)
		}

		_ = result
	},
}

// Helper functions for recursive operations

// uploadFile uploads a single file with progress bar
func uploadFile(ctx context.Context, source, dest string, size int64, backend storage.StorageBackend) {
	file, err := os.Open(source)
	if err != nil {
		color.Red("Error opening file: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	// Ensure destination has proper format (bucket/path)
	if !strings.Contains(dest, "/") {
		dest = dest + "/" + filepath.Base(source)
	} else if strings.HasSuffix(dest, "/") {
		dest = dest + filepath.Base(source)
	}

	// Progress bar
	bar := progressbar.DefaultBytes(size, "Uploading "+filepath.Base(source))

	opts := &storage.UploadOptions{
		ProgressFunc: func(bytes int64) {
			bar.Set64(bytes)
		},
	}

	result, err := backend.Upload(ctx, file, dest, opts)
	if err != nil {
		fmt.Println()
		color.Red("Error uploading %s: %v", filepath.Base(source), err)
		os.Exit(1)
	}

	fmt.Println()
	color.Green("✓ Upload complete: %s (%s)", filepath.Base(source), humanize.Bytes(uint64(result.Size)))
	fmt.Printf("  Location: %s\n", dest)
}

// uploadDir recursively uploads a directory
func uploadDir(ctx context.Context, source, dest string, backend storage.StorageBackend) {
	// Ensure dest ends with /
	if !strings.HasSuffix(dest, "/") {
		dest = dest + "/"
	}

	// Add base directory name to destination
	dest = dest + filepath.Base(source) + "/"

	var totalFiles int
	var totalSize int64

	// Walk the directory
	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		// Build remote path
		remotePath := dest + strings.ReplaceAll(relPath, "\\", "/")

		// Upload file
		uploadFile(ctx, path, remotePath, info.Size(), backend)

		totalFiles++
		totalSize += info.Size()

		return nil
	})

	if err != nil {
		color.Red("Error walking directory: %v", err)
		os.Exit(1)
	}

	fmt.Println()
	color.Green("✓ Directory upload complete!")
	fmt.Printf("  Files uploaded: %d\n", totalFiles)
	fmt.Printf("  Total size: %s\n", humanize.Bytes(uint64(totalSize)))
}

// downloadFile downloads a single file with progress bar
func downloadFile(ctx context.Context, source, dest string, size int64, backend storage.StorageBackend) {
	// Determine output filename
	outputPath := dest
	if info, err := os.Stat(dest); err == nil && info.IsDir() {
		outputPath = filepath.Join(dest, filepath.Base(source))
	}

	// Create output directory if needed
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		color.Red("Error creating directory: %v", err)
		os.Exit(1)
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		color.Red("Error creating file: %v", err)
		os.Exit(1)
	}
	defer outFile.Close()

	// Progress bar
	bar := progressbar.DefaultBytes(size, "Downloading "+filepath.Base(source))

	opts := &storage.DownloadOptions{
		ProgressFunc: func(bytes int64) {
			bar.Set64(bytes)
		},
	}

	result, err := backend.Download(ctx, source, outFile, opts)
	if err != nil {
		fmt.Println()
		color.Red("Error downloading %s: %v", filepath.Base(source), err)
		os.Exit(1)
	}

	fmt.Println()
	color.Green("✓ Download complete: %s (%s)", filepath.Base(outputPath), humanize.Bytes(uint64(result.Size)))
	fmt.Printf("  Saved to: %s\n", outputPath)
}

// downloadDir recursively downloads a directory
func downloadDir(ctx context.Context, source, dest string, backend storage.StorageBackend) {
	// List all files recursively
	opts := &storage.ListOptions{
		Recursive: true,
	}

	files, err := backend.List(ctx, source, opts)
	if err != nil {
		color.Red("Error listing files: %v", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		color.Yellow("No files found in %s", source)
		return
	}

	var totalFiles int
	var totalSize int64

	// Download each file
	for _, file := range files {
		if file.IsDir {
			continue
		}

		// Calculate local path
		relPath := strings.TrimPrefix(file.Path, source)
		relPath = strings.TrimPrefix(relPath, "/")
		localPath := filepath.Join(dest, relPath)

		// Download file
		downloadFile(ctx, file.Path, localPath, file.Size, backend)

		totalFiles++
		totalSize += file.Size
	}

	fmt.Println()
	color.Green("✓ Directory download complete!")
	fmt.Printf("  Files downloaded: %d\n", totalFiles)
	fmt.Printf("  Total size: %s\n", humanize.Bytes(uint64(totalSize)))
}

// deleteDir recursively deletes a directory
func deleteDir(ctx context.Context, path string, backend storage.StorageBackend) {
	// List all files recursively
	opts := &storage.ListOptions{
		Recursive: true,
	}

	files, err := backend.List(ctx, path, opts)
	if err != nil {
		color.Red("Error listing files: %v", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		color.Yellow("No files found in %s", path)
		return
	}

	var totalFiles int

	// Delete each file
	for _, file := range files {
		if file.IsDir {
			continue
		}

		if err := backend.Delete(ctx, file.Path); err != nil {
			color.Red("Error deleting %s: %v", file.Path, err)
			os.Exit(1)
		}

		color.Green("✓ Deleted: %s", file.Path)
		totalFiles++
	}

	fmt.Println()
	color.Green("✓ Directory delete complete!")
	fmt.Printf("  Files deleted: %d\n", totalFiles)
}

func init() {
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(mbCmd)
	rootCmd.AddCommand(putCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(cpCmd)
	rootCmd.AddCommand(mvCmd)
	rootCmd.AddCommand(catCmd)

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
}
