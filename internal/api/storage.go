package api

import (
	"fmt"
	"io"
	"os"
)

func (c *Client) UploadFile(localPath, remotePath string, progress UploadProgress) error {
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	var reader io.Reader = file
	if progress != nil {
		reader = &progressReader{
			reader:   file,
			progress: progress,
		}
	}

	fmt.Printf("Uploading %s to %s (%d bytes)\n", localPath, remotePath, stat.Size())
	_ = reader
	return nil
}

func (c *Client) DownloadFile(remotePath, localPath string, progress DownloadProgress) error {
	fmt.Printf("Downloading %s to %s\n", remotePath, localPath)
	return nil
}

func (c *Client) DeleteFile(remotePath string) error {
	fmt.Printf("Deleting %s\n", remotePath)
	return nil
}

func (c *Client) FileExists(remotePath string) (bool, error) {
	return false, nil
}

func (c *Client) GetFileHash(remotePath string) (string, error) {
	return "", nil
}
