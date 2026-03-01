package cmd

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var hashCmd = &cobra.Command{
	Use:   "hash <path>",
	Short: "Calculate file checksums",
	Long: `Calculate MD5, SHA256, and SHA512 checksums for a file.

Examples:
  darkstorage hash my-bucket/file.txt
  darkstorage hash my-bucket/file.txt --md5
  darkstorage hash my-bucket/file.txt --sha256`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		path := args[0]

		md5Only, _ := cmd.Flags().GetBool("md5")
		sha256Only, _ := cmd.Flags().GetBool("sha256")
		sha512Only, _ := cmd.Flags().GetBool("sha512")

		// If no specific hash selected, show all
		showAll := !md5Only && !sha256Only && !sha512Only

		// Create hashers
		hashers := make(map[string]hash.Hash)
		if showAll || md5Only {
			hashers["MD5"] = md5.New()
		}
		if showAll || sha256Only {
			hashers["SHA256"] = sha256.New()
		}
		if showAll || sha512Only {
			hashers["SHA512"] = sha512.New()
		}

		// Create multi-writer to hash while reading
		writers := make([]io.Writer, 0, len(hashers))
		for _, h := range hashers {
			writers = append(writers, h)
		}
		multiWriter := io.MultiWriter(writers...)

		// Download file and calculate hashes
		_, err := storageBackend.Download(ctx, path, multiWriter, nil)
		if err != nil {
			color.Red("Error downloading file: %v", err)
			os.Exit(1)
		}

		// Print results
		fmt.Printf("Checksums for %s:\n\n", path)
		if h, ok := hashers["MD5"]; ok {
			fmt.Printf("MD5:    %x\n", h.Sum(nil))
		}
		if h, ok := hashers["SHA256"]; ok {
			fmt.Printf("SHA256: %x\n", h.Sum(nil))
		}
		if h, ok := hashers["SHA512"]; ok {
			fmt.Printf("SHA512: %x\n", h.Sum(nil))
		}
	},
}

func init() {
	hashCmd.Flags().Bool("md5", false, "Calculate MD5 only")
	hashCmd.Flags().Bool("sha256", false, "Calculate SHA256 only")
	hashCmd.Flags().Bool("sha512", false, "Calculate SHA512 only")
	rootCmd.AddCommand(hashCmd)
}
