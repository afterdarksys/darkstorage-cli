package storage

import (
	"context"
	"io"
	"time"
)

// StorageBackend defines the interface that all storage backends must implement
// This allows us to support Traditional (MinIO/S3), Storj, IPFS, and Hybrid modes
type StorageBackend interface {
	// Basic operations
	Upload(ctx context.Context, src io.Reader, dest string, opts *UploadOptions) (*UploadResult, error)
	Download(ctx context.Context, src string, dest io.Writer, opts *DownloadOptions) (*DownloadResult, error)
	Delete(ctx context.Context, path string) error
	Copy(ctx context.Context, src, dest string) error
	Move(ctx context.Context, src, dest string) error

	// Listing and metadata
	List(ctx context.Context, prefix string, opts *ListOptions) ([]FileInfo, error)
	Stat(ctx context.Context, path string) (*FileInfo, error)

	// Bucket operations
	CreateBucket(ctx context.Context, name string) error
	DeleteBucket(ctx context.Context, name string) error
	ListBuckets(ctx context.Context) ([]BucketInfo, error)

	// Backend information
	BackendType() BackendType
	BackendInfo() map[string]interface{}

	// Health check
	Ping(ctx context.Context) error
}

// BackendType represents the storage backend type
type BackendType string

const (
	BackendTraditional BackendType = "traditional" // MinIO/S3
	BackendStorj       BackendType = "storj"       // Storj DCS
	BackendIPFS        BackendType = "ipfs"        // IPFS
	BackendHybrid      BackendType = "hybrid"      // Multiple backends
)

// StorageClass represents AWS-style storage tiers
type StorageClass string

const (
	StorageStandard          StorageClass = "STANDARD"           // Frequently accessed data
	StorageStandardIA        StorageClass = "STANDARD_IA"        // Infrequently accessed
	StorageIntelligentTiering StorageClass = "INTELLIGENT_TIERING" // Auto-tiering
	StorageGlacier           StorageClass = "GLACIER"            // Archive (3-5 hour retrieval)
	StorageGlacierInstant    StorageClass = "GLACIER_INSTANT"    // Archive (instant retrieval)
	StorageDeepArchive       StorageClass = "DEEP_ARCHIVE"       // Long-term (12 hour retrieval)
)

// StorageClassInfo provides metadata about a storage class
type StorageClassInfo struct {
	Class           StorageClass
	Name            string
	Description     string
	MinDuration     string // Minimum storage duration
	RetrievalTime   string // Typical retrieval time
	CostMultiplier  float64 // Relative cost (STANDARD = 1.0)
}

// GetStorageClassInfo returns metadata for a storage class
func GetStorageClassInfo(class StorageClass) StorageClassInfo {
	switch class {
	case StorageStandard:
		return StorageClassInfo{
			Class:          StorageStandard,
			Name:           "Standard",
			Description:    "For frequently accessed data",
			MinDuration:    "None",
			RetrievalTime:  "Instant",
			CostMultiplier: 1.0,
		}
	case StorageStandardIA:
		return StorageClassInfo{
			Class:          StorageStandardIA,
			Name:           "Standard Infrequent Access",
			Description:    "For infrequently accessed data",
			MinDuration:    "30 days",
			RetrievalTime:  "Instant",
			CostMultiplier: 0.5,
		}
	case StorageIntelligentTiering:
		return StorageClassInfo{
			Class:          StorageIntelligentTiering,
			Name:           "Intelligent Tiering",
			Description:    "Automatically moves data between tiers",
			MinDuration:    "None",
			RetrievalTime:  "Instant",
			CostMultiplier: 0.7,
		}
	case StorageGlacierInstant:
		return StorageClassInfo{
			Class:          StorageGlacierInstant,
			Name:           "Glacier Instant Retrieval",
			Description:    "Archive with instant retrieval",
			MinDuration:    "90 days",
			RetrievalTime:  "Instant",
			CostMultiplier: 0.3,
		}
	case StorageGlacier:
		return StorageClassInfo{
			Class:          StorageGlacier,
			Name:           "Glacier Flexible Retrieval",
			Description:    "Archive with 3-5 hour retrieval",
			MinDuration:    "90 days",
			RetrievalTime:  "3-5 hours",
			CostMultiplier: 0.1,
		}
	case StorageDeepArchive:
		return StorageClassInfo{
			Class:          StorageDeepArchive,
			Name:           "Deep Archive",
			Description:    "Long-term archive with 12 hour retrieval",
			MinDuration:    "180 days",
			RetrievalTime:  "12 hours",
			CostMultiplier: 0.04,
		}
	default:
		return StorageClassInfo{
			Class:          StorageStandard,
			Name:           "Standard",
			Description:    "Default storage class",
			MinDuration:    "None",
			RetrievalTime:  "Instant",
			CostMultiplier: 1.0,
		}
	}
}

// AllStorageClasses returns all available storage classes
func AllStorageClasses() []StorageClass {
	return []StorageClass{
		StorageStandard,
		StorageStandardIA,
		StorageIntelligentTiering,
		StorageGlacierInstant,
		StorageGlacier,
		StorageDeepArchive,
	}
}

// UploadOptions configures upload behavior
type UploadOptions struct {
	// Storage class
	StorageClass StorageClass

	// Content type (MIME type)
	ContentType string

	// Custom metadata
	Metadata map[string]string

	// Encryption (handled by encryption layer, not backend)
	// Encryption will be done before upload, so backend sees encrypted data

	// Progress callback
	ProgressFunc func(bytesTransferred int64)

	// Bandwidth limit (bytes per second, 0 = unlimited)
	BandwidthLimit int64

	// Part size for multipart upload (0 = auto-detect)
	PartSize int64

	// Number of concurrent parts for multipart upload
	ConcurrentParts int
}

// DownloadOptions configures download behavior
type DownloadOptions struct {
	// Progress callback
	ProgressFunc func(bytesTransferred int64)

	// Bandwidth limit (bytes per second, 0 = unlimited)
	BandwidthLimit int64

	// Resume from byte offset (for resumable downloads)
	ResumeFrom int64

	// Version ID (for versioned objects)
	VersionID string
}

// ListOptions configures list operations
type ListOptions struct {
	// Recursive listing
	Recursive bool

	// Prefix filter
	Prefix string

	// Max keys to return (0 = no limit)
	MaxKeys int

	// Start after this key (for pagination)
	StartAfter string

	// Include metadata
	IncludeMetadata bool
}

// UploadResult contains information about an upload
type UploadResult struct {
	Path          string
	Size          int64
	ETag          string
	VersionID     string
	StorageClass  StorageClass
	UploadedAt    time.Time
	BytesUploaded int64
	Duration      time.Duration
}

// DownloadResult contains information about a download
type DownloadResult struct {
	Path              string
	Size              int64
	ETag              string
	VersionID         string
	StorageClass      StorageClass
	LastModified      time.Time
	BytesDownloaded   int64
	Duration          time.Duration
}

// FileInfo represents file/object metadata
type FileInfo struct {
	Name         string
	Path         string
	Size         int64
	IsDir        bool
	ModifiedAt   time.Time
	ContentType  string
	ETag         string
	VersionID    string
	StorageClass StorageClass
	Metadata     map[string]string

	// Backend-specific info
	BackendType BackendType
	BackendData map[string]interface{}
}

// BucketInfo represents bucket metadata
type BucketInfo struct {
	Name        string
	CreatedAt   time.Time
	Region      string
	ObjectCount int64
	TotalSize   int64
}

// ProgressReader wraps io.Reader and reports progress
type ProgressReader struct {
	reader       io.Reader
	progressFunc func(int64)
	totalRead    int64
}

func NewProgressReader(r io.Reader, progressFunc func(int64)) *ProgressReader {
	return &ProgressReader{
		reader:       r,
		progressFunc: progressFunc,
	}
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.totalRead += int64(n)
	if pr.progressFunc != nil {
		pr.progressFunc(pr.totalRead)
	}
	return n, err
}

// BandwidthLimitedReader wraps io.Reader and limits bandwidth
type BandwidthLimitedReader struct {
	reader         io.Reader
	limit          int64 // bytes per second
	lastReadTime   time.Time
	bytesRead      int64
}

func NewBandwidthLimitedReader(r io.Reader, limit int64) *BandwidthLimitedReader {
	return &BandwidthLimitedReader{
		reader:       r,
		limit:        limit,
		lastReadTime: time.Now(),
	}
}

func (br *BandwidthLimitedReader) Read(p []byte) (int, error) {
	if br.limit <= 0 {
		return br.reader.Read(p)
	}

	// Calculate how long we should wait to maintain the limit
	elapsed := time.Since(br.lastReadTime)
	expectedDuration := time.Duration(float64(br.bytesRead) / float64(br.limit) * float64(time.Second))

	if expectedDuration > elapsed {
		time.Sleep(expectedDuration - elapsed)
	}

	n, err := br.reader.Read(p)
	br.bytesRead += int64(n)
	br.lastReadTime = time.Now()

	return n, err
}
