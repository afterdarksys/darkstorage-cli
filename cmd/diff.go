package cmd

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff <file1> <file2>",
	Short: "Compare two files",
	Long: `Compare two files and show differences.

Supports both text and binary files.

Examples:
  darkstorage diff my-bucket/file1.txt my-bucket/file2.txt
  darkstorage diff my-bucket/image1.png my-bucket/image2.png --binary
  darkstorage diff my-bucket/old.txt my-bucket/new.txt --unified`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		file1Path := args[0]
		file2Path := args[1]

		binary, _ := cmd.Flags().GetBool("binary")
		unified, _ := cmd.Flags().GetBool("unified")
		brief, _ := cmd.Flags().GetBool("brief")

		// Download both files
		var buf1, buf2 bytes.Buffer

		_, err := storageBackend.Download(ctx, file1Path, &buf1, nil)
		if err != nil {
			color.Red("Error downloading %s: %v", file1Path, err)
			os.Exit(1)
		}

		_, err = storageBackend.Download(ctx, file2Path, &buf2, nil)
		if err != nil {
			color.Red("Error downloading %s: %v", file2Path, err)
			os.Exit(1)
		}

		// Get file contents
		content1 := buf1.Bytes()
		content2 := buf2.Bytes()

		// Check if files are identical
		if bytes.Equal(content1, content2) {
			if !brief {
				color.Green("Files are identical")
			}
			return
		}

		fmt.Printf("Comparing: %s <=> %s\n\n", file1Path, file2Path)

		if brief {
			fmt.Println("Files differ")
			return
		}

		// Binary diff
		if binary || !isText(content1) || !isText(content2) {
			binaryDiff(content1, content2)
			return
		}

		// Text diff
		if unified {
			unifiedDiff(string(content1), string(content2), file1Path, file2Path)
		} else {
			textDiff(string(content1), string(content2))
		}
	},
}

func isText(data []byte) bool {
	// Simple heuristic: check for null bytes
	for _, b := range data {
		if b == 0 {
			return false
		}
	}
	return true
}

func textDiff(text1, text2 string) {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(text1, text2, false)

	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			color.Green("+ %s", diff.Text)
		case diffmatchpatch.DiffDelete:
			color.Red("- %s", diff.Text)
		case diffmatchpatch.DiffEqual:
			// Show context
			lines := strings.Split(diff.Text, "\n")
			if len(lines) > 3 {
				// Show first and last line with ...
				fmt.Printf("  %s\n", lines[0])
				if len(lines) > 4 {
					fmt.Println("  ...")
				}
				fmt.Printf("  %s\n", lines[len(lines)-1])
			} else {
				for _, line := range lines {
					fmt.Printf("  %s\n", line)
				}
			}
		}
	}
}

func unifiedDiff(text1, text2, file1, file2 string) {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(text1, text2, false)
	patches := dmp.PatchMake(text1, diffs)

	fmt.Printf("--- %s\n", file1)
	fmt.Printf("+++ %s\n", file2)

	for _, patch := range patches {
		fmt.Println(patch)
	}
}

func binaryDiff(data1, data2 []byte) {
	fmt.Println("Binary files - showing hex diff:\n")

	maxLen := len(data1)
	if len(data2) > maxLen {
		maxLen = len(data2)
	}

	// Limit output to first 256 bytes
	if maxLen > 256 {
		maxLen = 256
		fmt.Println("(Showing first 256 bytes only)\n")
	}

	differences := 0
	for i := 0; i < maxLen; i += 16 {
		end := i + 16
		if end > maxLen {
			end = maxLen
		}

		chunk1 := make([]byte, 16)
		chunk2 := make([]byte, 16)

		// Fill chunks
		for j := i; j < end; j++ {
			if j < len(data1) {
				chunk1[j-i] = data1[j]
			}
			if j < len(data2) {
				chunk2[j-i] = data2[j]
			}
		}

		// Check if chunks differ
		differ := !bytes.Equal(chunk1[:end-i], chunk2[:end-i])

		if differ {
			differences++
			fmt.Printf("%08x: ", i)

			// Show hex for file1
			fmt.Printf("%-48s  ", hex.EncodeToString(chunk1[:end-i]))

			// Show hex for file2
			color.Yellow("%-48s\n", hex.EncodeToString(chunk2[:end-i]))
		}
	}

	fmt.Printf("\nSize: %d vs %d bytes\n", len(data1), len(data2))
	fmt.Printf("Differences: %d chunks (16 bytes each)\n", differences)

	if len(data1) != len(data2) {
		color.Yellow("Warning: Files have different sizes")
	}
}

func init() {
	diffCmd.Flags().Bool("binary", false, "Force binary diff mode")
	diffCmd.Flags().Bool("unified", false, "Use unified diff format (text files)")
	diffCmd.Flags().Bool("brief", false, "Only show if files differ, not the differences")

	rootCmd.AddCommand(diffCmd)
}
