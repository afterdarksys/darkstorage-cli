package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var fileCmd = &cobra.Command{
	Use:   "file <path>",
	Short: "Detect file type",
	Long: `Detect the MIME type and file format of a stored file.

Examples:
  darkstorage file my-bucket/image.jpg
  darkstorage file my-bucket/document.pdf`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		path := args[0]

		// Download first 512 bytes for type detection
		buffer := make([]byte, 512)
		bufWriter := &limitedWriter{buf: buffer}

		_, err := storageBackend.Download(ctx, path, bufWriter, nil)
		if err != nil {
			color.Red("Error downloading file: %v", err)
			os.Exit(1)
		}

		n := bufWriter.n

		// Detect content type
		contentType := http.DetectContentType(buffer[:n])

		// Get file info for size
		// Note: This would require adding a Stat method to StorageBackend
		// For now, just show the MIME type

		fmt.Printf("%s: %s\n", path, contentType)

		// Show additional details based on type
		switch contentType {
		case "application/zip":
			fmt.Println("  Type: ZIP archive")
		case "application/x-gzip":
			fmt.Println("  Type: Gzip compressed")
		case "application/x-tar":
			fmt.Println("  Type: TAR archive")
		case "application/pdf":
			fmt.Println("  Type: PDF document")
		case "text/plain; charset=utf-8":
			fmt.Println("  Type: Text file")
		case "application/octet-stream":
			fmt.Println("  Type: Binary data")
		}

		// Detect additional formats by magic bytes
		if len(buffer) >= 4 {
			// Check for common archive formats
			if buffer[0] == 0x50 && buffer[1] == 0x4B && buffer[2] == 0x03 && buffer[3] == 0x04 {
				fmt.Println("  Format: ZIP/JAR/APK/DOCX/XLSX")
			} else if buffer[0] == 0x1F && buffer[1] == 0x8B {
				fmt.Println("  Format: GZIP")
			} else if buffer[0] == 0x42 && buffer[1] == 0x5A && buffer[2] == 0x68 {
				fmt.Println("  Format: BZIP2")
			} else if buffer[0] == 'u' && buffer[1] == 's' && buffer[2] == 't' && buffer[3] == 'a' {
				if len(buffer) >= 263 && buffer[257] == 'u' && buffer[258] == 's' && buffer[259] == 't' && buffer[260] == 'a' && buffer[261] == 'r' {
					fmt.Println("  Format: TAR archive")
				}
			}
		}
	},
}

// limitedWriter writes to a buffer up to its capacity
type limitedWriter struct {
	buf []byte
	n   int
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	remaining := len(w.buf) - w.n
	if remaining <= 0 {
		return len(p), nil // Discard excess data
	}

	toCopy := len(p)
	if toCopy > remaining {
		toCopy = remaining
	}

	copy(w.buf[w.n:], p[:toCopy])
	w.n += toCopy
	return len(p), nil // Always return full length to avoid errors
}

func init() {
	rootCmd.AddCommand(fileCmd)
}
