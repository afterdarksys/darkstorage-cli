package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// TraditionalBackend implements StorageBackend for MinIO/S3
type TraditionalBackend struct {
	client   *minio.Client
	endpoint string
	useSSL   bool
	region   string
}

// TraditionalConfig contains configuration for Traditional (MinIO/S3) backend
type TraditionalConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Region    string
}

// NewTraditionalBackend creates a new Traditional storage backend
func NewTraditionalBackend(cfg *TraditionalConfig) (*TraditionalBackend, error) {
	// Create MinIO client
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	return &TraditionalBackend{
		client:   client,
		endpoint: cfg.Endpoint,
		useSSL:   cfg.UseSSL,
		region:   cfg.Region,
	}, nil
}

// parsePath splits "bucket/path/to/file.txt" into bucket and object path
func parsePath(path string) (bucket, object string) {
	path = strings.TrimPrefix(path, "/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

// Upload uploads a file to MinIO/S3
func (t *TraditionalBackend) Upload(ctx context.Context, src io.Reader, dest string, opts *UploadOptions) (*UploadResult, error) {
	bucket, object := parsePath(dest)
	if object == "" {
		return nil, fmt.Errorf("invalid destination path: %s (must be bucket/object)", dest)
	}

	// Default options
	if opts == nil {
		opts = &UploadOptions{}
	}

	// Wrap reader with progress and bandwidth limiting
	var reader io.Reader = src
	if opts.ProgressFunc != nil {
		reader = NewProgressReader(reader, opts.ProgressFunc)
	}
	if opts.BandwidthLimit > 0 {
		reader = NewBandwidthLimitedReader(reader, opts.BandwidthLimit)
	}

	// Prepare put options
	putOpts := minio.PutObjectOptions{
		ContentType: opts.ContentType,
		UserMetadata: opts.Metadata,
	}

	// Set storage class if specified
	if opts.StorageClass != "" {
		putOpts.StorageClass = string(opts.StorageClass)
	}

	// Set part size for multipart upload
	if opts.PartSize > 0 {
		putOpts.PartSize = uint64(opts.PartSize)
	}

	// Upload
	startTime := time.Now()
	info, err := t.client.PutObject(ctx, bucket, object, reader, -1, putOpts)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	return &UploadResult{
		Path:          dest,
		Size:          info.Size,
		ETag:          info.ETag,
		VersionID:     info.VersionID,
		StorageClass:  opts.StorageClass,
		UploadedAt:    time.Now(),
		BytesUploaded: info.Size,
		Duration:      time.Since(startTime),
	}, nil
}

// Download downloads a file from MinIO/S3
func (t *TraditionalBackend) Download(ctx context.Context, src string, dest io.Writer, opts *DownloadOptions) (*DownloadResult, error) {
	bucket, object := parsePath(src)
	if object == "" {
		return nil, fmt.Errorf("invalid source path: %s (must be bucket/object)", src)
	}

	// Default options
	if opts == nil {
		opts = &DownloadOptions{}
	}

	// Prepare get options
	getOpts := minio.GetObjectOptions{}
	if opts.VersionID != "" {
		getOpts.VersionID = opts.VersionID
	}

	// Get object
	startTime := time.Now()
	obj, err := t.client.GetObject(ctx, bucket, object, getOpts)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer obj.Close()

	// Get object info
	stat, err := obj.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get object info: %w", err)
	}

	// Wrap reader with progress and bandwidth limiting
	var reader io.Reader = obj
	if opts.ProgressFunc != nil {
		reader = NewProgressReader(reader, opts.ProgressFunc)
	}
	if opts.BandwidthLimit > 0 {
		reader = NewBandwidthLimitedReader(reader, opts.BandwidthLimit)
	}

	// Copy to destination
	bytesWritten, err := io.Copy(dest, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to write download: %w", err)
	}

	return &DownloadResult{
		Path:            src,
		Size:            bytesWritten,
		ETag:            stat.ETag,
		VersionID:       stat.VersionID,
		StorageClass:    StorageClass(stat.StorageClass),
		LastModified:    stat.LastModified,
		BytesDownloaded: bytesWritten,
		Duration:        time.Since(startTime),
	}, nil
}

// Delete deletes a file from MinIO/S3
func (t *TraditionalBackend) Delete(ctx context.Context, path string) error {
	bucket, object := parsePath(path)
	if object == "" {
		return fmt.Errorf("invalid path: %s (must be bucket/object)", path)
	}

	err := t.client.RemoveObject(ctx, bucket, object, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	return nil
}

// Copy copies a file within MinIO/S3
func (t *TraditionalBackend) Copy(ctx context.Context, src, dest string) error {
	srcBucket, srcObject := parsePath(src)
	destBucket, destObject := parsePath(dest)

	if srcObject == "" || destObject == "" {
		return fmt.Errorf("invalid paths (must be bucket/object)")
	}

	// Prepare copy source
	srcOpts := minio.CopySrcOptions{
		Bucket: srcBucket,
		Object: srcObject,
	}

	// Copy
	_, err := t.client.CopyObject(ctx, minio.CopyDestOptions{
		Bucket: destBucket,
		Object: destObject,
	}, srcOpts)

	if err != nil {
		return fmt.Errorf("copy failed: %w", err)
	}

	return nil
}

// Move moves a file within MinIO/S3 (copy + delete)
func (t *TraditionalBackend) Move(ctx context.Context, src, dest string) error {
	// Copy first
	if err := t.Copy(ctx, src, dest); err != nil {
		return err
	}

	// Delete source
	if err := t.Delete(ctx, src); err != nil {
		return fmt.Errorf("move failed (copied but could not delete source): %w", err)
	}

	return nil
}

// List lists files in MinIO/S3
func (t *TraditionalBackend) List(ctx context.Context, prefix string, opts *ListOptions) ([]FileInfo, error) {
	bucket, objectPrefix := parsePath(prefix)

	// Default options
	if opts == nil {
		opts = &ListOptions{}
	}

	// List objects
	listOpts := minio.ListObjectsOptions{
		Prefix:    objectPrefix,
		Recursive: opts.Recursive,
	}

	var files []FileInfo
	for obj := range t.client.ListObjects(ctx, bucket, listOpts) {
		if obj.Err != nil {
			return nil, fmt.Errorf("list failed: %w", obj.Err)
		}

		files = append(files, FileInfo{
			Name:         filepath.Base(obj.Key),
			Path:         bucket + "/" + obj.Key,
			Size:         obj.Size,
			IsDir:        strings.HasSuffix(obj.Key, "/"),
			ModifiedAt:   obj.LastModified,
			ContentType:  obj.ContentType,
			ETag:         obj.ETag,
			VersionID:    obj.VersionID,
			StorageClass: StorageClass(obj.StorageClass),
			Metadata:     obj.UserMetadata,
			BackendType:  BackendTraditional,
		})

		// Respect MaxKeys limit
		if opts.MaxKeys > 0 && len(files) >= opts.MaxKeys {
			break
		}
	}

	return files, nil
}

// Stat gets metadata for a file
func (t *TraditionalBackend) Stat(ctx context.Context, path string) (*FileInfo, error) {
	bucket, object := parsePath(path)
	if object == "" {
		return nil, fmt.Errorf("invalid path: %s (must be bucket/object)", path)
	}

	info, err := t.client.StatObject(ctx, bucket, object, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("stat failed: %w", err)
	}

	return &FileInfo{
		Name:         filepath.Base(object),
		Path:         path,
		Size:         info.Size,
		IsDir:        false,
		ModifiedAt:   info.LastModified,
		ContentType:  info.ContentType,
		ETag:         info.ETag,
		VersionID:    info.VersionID,
		StorageClass: StorageClass(info.StorageClass),
		Metadata:     info.UserMetadata,
		BackendType:  BackendTraditional,
	}, nil
}

// CreateBucket creates a new bucket
func (t *TraditionalBackend) CreateBucket(ctx context.Context, name string) error {
	err := t.client.MakeBucket(ctx, name, minio.MakeBucketOptions{
		Region: t.region,
	})
	if err != nil {
		return fmt.Errorf("create bucket failed: %w", err)
	}
	return nil
}

// DeleteBucket deletes a bucket (must be empty)
func (t *TraditionalBackend) DeleteBucket(ctx context.Context, name string) error {
	err := t.client.RemoveBucket(ctx, name)
	if err != nil {
		return fmt.Errorf("delete bucket failed: %w", err)
	}
	return nil
}

// ListBuckets lists all buckets
func (t *TraditionalBackend) ListBuckets(ctx context.Context) ([]BucketInfo, error) {
	buckets, err := t.client.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("list buckets failed: %w", err)
	}

	var result []BucketInfo
	for _, bucket := range buckets {
		result = append(result, BucketInfo{
			Name:      bucket.Name,
			CreatedAt: bucket.CreationDate,
		})
	}

	return result, nil
}

// BackendType returns the backend type
func (t *TraditionalBackend) BackendType() BackendType {
	return BackendTraditional
}

// BackendInfo returns backend-specific information
func (t *TraditionalBackend) BackendInfo() map[string]interface{} {
	return map[string]interface{}{
		"type":     "traditional",
		"endpoint": t.endpoint,
		"ssl":      t.useSSL,
		"region":   t.region,
		"provider": "MinIO/S3",
	}
}

// Ping checks if the backend is accessible
func (t *TraditionalBackend) Ping(ctx context.Context) error {
	_, err := t.client.ListBuckets(ctx)
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	return nil
}
