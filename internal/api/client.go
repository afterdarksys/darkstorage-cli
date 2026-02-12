package api

import (
	"io"
	"net/http"
	"time"
)

type Client struct {
	endpoint   string
	apiKey     string
	httpClient *http.Client
	timeout    time.Duration
}

func NewClient(endpoint, apiKey string) *Client {
	return &Client{
		endpoint: endpoint,
		apiKey:   apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		timeout: 30 * time.Second,
	}
}

func (c *Client) SetTimeout(duration time.Duration) {
	c.timeout = duration
	c.httpClient.Timeout = duration
}

type UploadProgress func(bytesTransferred int64)
type DownloadProgress func(bytesTransferred int64)

type progressReader struct {
	reader   io.Reader
	progress UploadProgress
	total    int64
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.total += int64(n)
	if pr.progress != nil {
		pr.progress(pr.total)
	}
	return n, err
}
