package cmd

import (
	"compress/bzip2"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/ulikunitz/xz"
)

// gzCmd - compress files with gzip
var gzCmd = &cobra.Command{
	Use:   "gz <source> [destination]",
	Short: "Compress file with gzip",
	Long: `Compress a file using gzip compression.

Examples:
  darkstorage gz my-bucket/file.txt
  darkstorage gz my-bucket/file.txt my-bucket/file.txt.gz`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		source := args[0]
		dest := source + ".gz"
		if len(args) > 1 {
			dest = args[1]
		}

		level, _ := cmd.Flags().GetInt("level")

		if err := compressGzip(ctx, source, dest, level); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		color.Green("✓ Compressed: %s → %s", source, dest)
	},
}

// bz2Cmd - compress files with bzip2
var bz2Cmd = &cobra.Command{
	Use:   "bz2 <source> [destination]",
	Short: "Compress file with bzip2",
	Long: `Compress a file using bzip2 compression.

Examples:
  darkstorage bz2 my-bucket/file.txt
  darkstorage bz2 my-bucket/file.txt my-bucket/file.txt.bz2`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		source := args[0]
		dest := source + ".bz2"
		if len(args) > 1 {
			dest = args[1]
		}

		if err := compressBzip2(ctx, source, dest); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		color.Green("✓ Compressed: %s → %s", source, dest)
	},
}

// xzCmd - compress files with xz
var xzCmd = &cobra.Command{
	Use:   "xz <source> [destination]",
	Short: "Compress file with xz (LZMA2)",
	Long: `Compress a file using xz compression (LZMA2 algorithm).

Examples:
  darkstorage xz my-bucket/file.txt
  darkstorage xz my-bucket/file.txt my-bucket/file.txt.xz`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		source := args[0]
		dest := source + ".xz"
		if len(args) > 1 {
			dest = args[1]
		}

		if err := compressXZ(ctx, source, dest); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		color.Green("✓ Compressed: %s → %s", source, dest)
	},
}

// gunzipCmd - decompress gzip files
var gunzipCmd = &cobra.Command{
	Use:   "gunzip <source> [destination]",
	Short: "Decompress gzip file",
	Long: `Decompress a gzip compressed file.

Examples:
  darkstorage gunzip my-bucket/file.txt.gz
  darkstorage gunzip my-bucket/file.txt.gz my-bucket/file.txt`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		source := args[0]
		dest := strings.TrimSuffix(source, ".gz")
		if len(args) > 1 {
			dest = args[1]
		}

		if err := decompressGzip(ctx, source, dest); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		color.Green("✓ Decompressed: %s → %s", source, dest)
	},
}

// bunzip2Cmd - decompress bzip2 files
var bunzip2Cmd = &cobra.Command{
	Use:   "bunzip2 <source> [destination]",
	Short: "Decompress bzip2 file",
	Long: `Decompress a bzip2 compressed file.

Examples:
  darkstorage bunzip2 my-bucket/file.txt.bz2
  darkstorage bunzip2 my-bucket/file.txt.bz2 my-bucket/file.txt`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		source := args[0]
		dest := strings.TrimSuffix(source, ".bz2")
		if len(args) > 1 {
			dest = args[1]
		}

		if err := decompressBzip2(ctx, source, dest); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		color.Green("✓ Decompressed: %s → %s", source, dest)
	},
}

// unxzCmd - decompress xz files
var unxzCmd = &cobra.Command{
	Use:   "unxz <source> [destination]",
	Short: "Decompress xz file",
	Long: `Decompress an xz compressed file.

Examples:
  darkstorage unxz my-bucket/file.txt.xz
  darkstorage unxz my-bucket/file.txt.xz my-bucket/file.txt`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		source := args[0]
		dest := strings.TrimSuffix(source, ".xz")
		if len(args) > 1 {
			dest = args[1]
		}

		if err := decompressXZ(ctx, source, dest); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		color.Green("✓ Decompressed: %s → %s", source, dest)
	},
}

// Compression functions

func compressGzip(ctx context.Context, source, dest string, level int) error {
	// Create temporary file for compression
	tmpFile, err := os.CreateTemp("", "darkstorage-gz-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Create gzip writer
	var gzWriter *gzip.Writer
	if level >= 0 && level <= 9 {
		gzWriter, err = gzip.NewWriterLevel(tmpFile, level)
	} else {
		gzWriter = gzip.NewWriter(tmpFile)
	}
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to create gzip writer: %w", err)
	}

	// Set gzip header
	gzWriter.Name = filepath.Base(source)

	// Download and compress
	_, err = storageBackend.Download(ctx, source, gzWriter, nil)
	if err != nil {
		gzWriter.Close()
		tmpFile.Close()
		return fmt.Errorf("failed to download file: %w", err)
	}

	if err := gzWriter.Close(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Upload compressed file
	compressedFile, err := os.Open(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to open compressed file: %w", err)
	}
	defer compressedFile.Close()

	_, err = storageBackend.Upload(ctx, compressedFile, dest, nil)
	if err != nil {
		return fmt.Errorf("failed to upload compressed file: %w", err)
	}

	return nil
}

func compressBzip2(ctx context.Context, source, dest string) error {
	color.Red("BZIP2 compression not yet implemented")
	color.Yellow("Note: Go's standard library only supports bzip2 decompression, not compression")
	color.Yellow("Consider using gzip (gz) or xz instead")
	return fmt.Errorf("bzip2 compression requires external package")
}

func compressXZ(ctx context.Context, source, dest string) error {
	// Create temporary file for compression
	tmpFile, err := os.CreateTemp("", "darkstorage-xz-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Create xz writer
	xzWriter, err := xz.NewWriter(tmpFile)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to create xz writer: %w", err)
	}

	// Download and compress
	_, err = storageBackend.Download(ctx, source, xzWriter, nil)
	if err != nil {
		xzWriter.Close()
		tmpFile.Close()
		return fmt.Errorf("failed to download file: %w", err)
	}

	if err := xzWriter.Close(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to close xz writer: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Upload compressed file
	compressedFile, err := os.Open(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to open compressed file: %w", err)
	}
	defer compressedFile.Close()

	_, err = storageBackend.Upload(ctx, compressedFile, dest, nil)
	if err != nil {
		return fmt.Errorf("failed to upload compressed file: %w", err)
	}

	return nil
}

// Decompression functions

func decompressGzip(ctx context.Context, source, dest string) error {
	// Create temporary file for decompression
	tmpFile, err := os.CreateTemp("", "darkstorage-gunzip-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Download compressed file
	_, err = storageBackend.Download(ctx, source, tmpFile, nil)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to download file: %w", err)
	}
	tmpFile.Close()

	// Open and decompress
	tmpFile, err = os.Open(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to open temp file: %w", err)
	}
	defer tmpFile.Close()

	gzReader, err := gzip.NewReader(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create output temp file
	outFile, err := os.CreateTemp("", "darkstorage-out-*")
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	outPath := outFile.Name()
	defer os.Remove(outPath)

	_, err = io.Copy(outFile, gzReader)
	if err != nil {
		outFile.Close()
		return fmt.Errorf("failed to decompress: %w", err)
	}
	outFile.Close()

	// Upload decompressed file
	decompressedFile, err := os.Open(outPath)
	if err != nil {
		return fmt.Errorf("failed to open decompressed file: %w", err)
	}
	defer decompressedFile.Close()

	_, err = storageBackend.Upload(ctx, decompressedFile, dest, nil)
	if err != nil {
		return fmt.Errorf("failed to upload decompressed file: %w", err)
	}

	return nil
}

func decompressBzip2(ctx context.Context, source, dest string) error {
	// Create temporary file for decompression
	tmpFile, err := os.CreateTemp("", "darkstorage-bunzip2-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Download compressed file
	_, err = storageBackend.Download(ctx, source, tmpFile, nil)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to download file: %w", err)
	}
	tmpFile.Close()

	// Open and decompress
	tmpFile, err = os.Open(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to open temp file: %w", err)
	}
	defer tmpFile.Close()

	bz2Reader := bzip2.NewReader(tmpFile)

	// Create output temp file
	outFile, err := os.CreateTemp("", "darkstorage-out-*")
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	outPath := outFile.Name()
	defer os.Remove(outPath)

	_, err = io.Copy(outFile, bz2Reader)
	if err != nil {
		outFile.Close()
		return fmt.Errorf("failed to decompress: %w", err)
	}
	outFile.Close()

	// Upload decompressed file
	decompressedFile, err := os.Open(outPath)
	if err != nil {
		return fmt.Errorf("failed to open decompressed file: %w", err)
	}
	defer decompressedFile.Close()

	_, err = storageBackend.Upload(ctx, decompressedFile, dest, nil)
	if err != nil {
		return fmt.Errorf("failed to upload decompressed file: %w", err)
	}

	return nil
}

func decompressXZ(ctx context.Context, source, dest string) error {
	// Create temporary file for decompression
	tmpFile, err := os.CreateTemp("", "darkstorage-unxz-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Download compressed file
	_, err = storageBackend.Download(ctx, source, tmpFile, nil)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to download file: %w", err)
	}
	tmpFile.Close()

	// Open and decompress
	tmpFile, err = os.Open(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to open temp file: %w", err)
	}
	defer tmpFile.Close()

	xzReader, err := xz.NewReader(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create xz reader: %w", err)
	}

	// Create output temp file
	outFile, err := os.CreateTemp("", "darkstorage-out-*")
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	outPath := outFile.Name()
	defer os.Remove(outPath)

	_, err = io.Copy(outFile, xzReader)
	if err != nil {
		outFile.Close()
		return fmt.Errorf("failed to decompress: %w", err)
	}
	outFile.Close()

	// Upload decompressed file
	decompressedFile, err := os.Open(outPath)
	if err != nil {
		return fmt.Errorf("failed to open decompressed file: %w", err)
	}
	defer decompressedFile.Close()

	_, err = storageBackend.Upload(ctx, decompressedFile, dest, nil)
	if err != nil {
		return fmt.Errorf("failed to upload decompressed file: %w", err)
	}

	return nil
}

func init() {
	gzCmd.Flags().IntP("level", "l", -1, "Compression level (1-9, default: 6)")

	rootCmd.AddCommand(gzCmd)
	rootCmd.AddCommand(bz2Cmd)
	rootCmd.AddCommand(xzCmd)
	rootCmd.AddCommand(gunzipCmd)
	rootCmd.AddCommand(bunzip2Cmd)
	rootCmd.AddCommand(unxzCmd)
}
