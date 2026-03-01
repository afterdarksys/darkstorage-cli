package cmd

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

// zip command
var zipCmd = &cobra.Command{
	Use:   "zip <archive-name> <files...>",
	Short: "Create a ZIP archive",
	Long: `Create a ZIP archive from one or more files in storage.

Examples:
  darkstorage zip archive.zip my-bucket/file1.txt my-bucket/file2.txt
  darkstorage zip backup.zip my-bucket/folder/ --recursive`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		archiveName := args[0]
		files := args[1:]
		recursive, _ := cmd.Flags().GetBool("recursive")

		// Create local ZIP file
		zipFile, err := os.Create(archiveName)
		if err != nil {
			color.Red("Error creating archive: %v", err)
			os.Exit(1)
		}
		defer zipFile.Close()

		zipWriter := zip.NewWriter(zipFile)
		defer zipWriter.Close()

		color.Green("Creating ZIP archive: %s", archiveName)

		for _, path := range files {
			if err := addToZip(ctx, zipWriter, path, recursive); err != nil {
				color.Red("Error adding %s: %v", path, err)
				os.Exit(1)
			}
		}

		color.Green("✓ Archive created: %s", archiveName)
	},
}

// tar command
var tarCmd = &cobra.Command{
	Use:   "tar <archive-name> <files...>",
	Short: "Create a TAR archive",
	Long: `Create a TAR, TAR.GZ, or TAR.BZ2 archive from files in storage.

Examples:
  darkstorage tar archive.tar my-bucket/file1.txt
  darkstorage tar backup.tar.gz my-bucket/folder/ --gzip
  darkstorage tar backup.tar.bz2 my-bucket/folder/ --bzip2`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		archiveName := args[0]
		files := args[1:]
		useGzip, _ := cmd.Flags().GetBool("gzip")
		useBzip2, _ := cmd.Flags().GetBool("bzip2")
		recursive, _ := cmd.Flags().GetBool("recursive")

		// Create local TAR file
		tarFile, err := os.Create(archiveName)
		if err != nil {
			color.Red("Error creating archive: %v", err)
			os.Exit(1)
		}
		defer tarFile.Close()

		var tarWriter *tar.Writer

		if useGzip {
			gzipWriter := gzip.NewWriter(tarFile)
			defer gzipWriter.Close()
			tarWriter = tar.NewWriter(gzipWriter)
		} else if useBzip2 {
			// Note: bzip2 writer requires external package
			color.Red("BZIP2 compression not yet implemented")
			os.Exit(1)
		} else {
			tarWriter = tar.NewWriter(tarFile)
		}
		defer tarWriter.Close()

		color.Green("Creating TAR archive: %s", archiveName)

		for _, path := range files {
			if err := addToTar(ctx, tarWriter, path, recursive); err != nil {
				color.Red("Error adding %s: %v", path, err)
				os.Exit(1)
			}
		}

		color.Green("✓ Archive created: %s", archiveName)
	},
}

// extract command
var extractCmd = &cobra.Command{
	Use:   "extract <archive> [destination-bucket/path]",
	Short: "Extract an archive to storage",
	Long: `Extract ZIP, TAR, TAR.GZ, or TAR.BZ2 archives to storage.

Examples:
  darkstorage extract archive.zip my-bucket/
  darkstorage extract backup.tar.gz my-bucket/restored/`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initStorage(); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		archivePath := args[0]
		destPath := ""
		if len(args) > 1 {
			destPath = args[1]
		}

		// Determine archive type
		if strings.HasSuffix(archivePath, ".zip") {
			if err := extractZip(ctx, archivePath, destPath); err != nil {
				color.Red("Error extracting ZIP: %v", err)
				os.Exit(1)
			}
		} else if strings.HasSuffix(archivePath, ".tar") {
			if err := extractTar(ctx, archivePath, destPath, false, false); err != nil {
				color.Red("Error extracting TAR: %v", err)
				os.Exit(1)
			}
		} else if strings.HasSuffix(archivePath, ".tar.gz") || strings.HasSuffix(archivePath, ".tgz") {
			if err := extractTar(ctx, archivePath, destPath, true, false); err != nil {
				color.Red("Error extracting TAR.GZ: %v", err)
				os.Exit(1)
			}
		} else if strings.HasSuffix(archivePath, ".tar.bz2") || strings.HasSuffix(archivePath, ".tbz2") {
			if err := extractTar(ctx, archivePath, destPath, false, true); err != nil {
				color.Red("Error extracting TAR.BZ2: %v", err)
				os.Exit(1)
			}
		} else {
			color.Red("Unsupported archive format. Supported: .zip, .tar, .tar.gz, .tgz, .tar.bz2, .tbz2")
			os.Exit(1)
		}

		color.Green("✓ Archive extracted successfully")
	},
}

// Helper functions
func addToZip(ctx context.Context, zipWriter *zip.Writer, path string, recursive bool) error {
	// Get filename from path
	filename := filepath.Base(path)

	writer, err := zipWriter.Create(filename)
	if err != nil {
		return err
	}

	_, err = storageBackend.Download(ctx, path, writer, nil)
	return err
}

func addToTar(ctx context.Context, tarWriter *tar.Writer, path string, recursive bool) error {
	// Get filename from path
	filename := filepath.Base(path)

	// Note: We'd need file size for tar header
	// For now, read into memory (not ideal for large files)
	var buf bytes.Buffer
	_, err := storageBackend.Download(ctx, path, &buf, nil)
	if err != nil {
		return err
	}

	data := buf.Bytes()

	header := &tar.Header{
		Name: filename,
		Mode: 0644,
		Size: int64(len(data)),
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}

	_, err = tarWriter.Write(data)
	return err
}

func extractZip(ctx context.Context, archivePath, destPath string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	bar := progressbar.Default(int64(len(reader.File)), "Extracting")

	for _, file := range reader.File {
		if err := extractZipFile(ctx, file, destPath); err != nil {
			return err
		}
		bar.Add(1)
	}

	return nil
}

func extractZipFile(ctx context.Context, file *zip.File, destPath string) error {
	reader, err := file.Open()
	if err != nil {
		return err
	}
	defer reader.Close()

	targetPath := filepath.Join(destPath, file.Name)
	_, err = storageBackend.Upload(ctx, reader, targetPath, nil)
	return err
}

func extractTar(ctx context.Context, archivePath, destPath string, useGzip, useBzip2 bool) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var tarReader *tar.Reader

	if useGzip {
		gzipReader, err := gzip.NewReader(file)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		tarReader = tar.NewReader(gzipReader)
	} else if useBzip2 {
		bzip2Reader := bzip2.NewReader(file)
		tarReader = tar.NewReader(bzip2Reader)
	} else {
		tarReader = tar.NewReader(file)
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if header.Typeflag == tar.TypeReg {
			targetPath := filepath.Join(destPath, header.Name)
			if _, err := storageBackend.Upload(ctx, tarReader, targetPath, nil); err != nil {
				return err
			}
		}
	}

	return nil
}

func init() {
	zipCmd.Flags().Bool("recursive", false, "Include subdirectories")
	tarCmd.Flags().Bool("recursive", false, "Include subdirectories")
	tarCmd.Flags().Bool("gzip", false, "Use GZIP compression (TAR.GZ)")
	tarCmd.Flags().Bool("bzip2", false, "Use BZIP2 compression (TAR.BZ2)")

	rootCmd.AddCommand(zipCmd)
	rootCmd.AddCommand(tarCmd)
	rootCmd.AddCommand(extractCmd)
}
