# Dark Storage GUI & Daemon - Data Structures Reference

## Core Data Structures

All structs should be defined in their respective packages. This document serves as a reference.

### Configuration Structures

```go
// internal/config/daemon.go
package config

import "time"

type DaemonConfig struct {
    Daemon       DaemonSettings       `yaml:"daemon"`
    SyncFolders  []SyncFolderConfig   `yaml:"sync_folders"`
    Notifications NotificationSettings `yaml:"notifications"`
    API          APIConfig            `yaml:"api"`
}

type DaemonSettings struct {
    Enabled         bool          `yaml:"enabled"`
    LogLevel        string        `yaml:"log_level"`
    LogFile         string        `yaml:"log_file"`
    PIDFile         string        `yaml:"pid_file"`
    IPCSocket       string        `yaml:"ipc_socket"`
    WorkerThreads   int           `yaml:"worker_threads"`
    MaxQueueSize    int           `yaml:"max_queue_size"`
    DebounceDelay   time.Duration `yaml:"debounce_delay"`
    Timeout         time.Duration `yaml:"timeout"`
    RetryAttempts   int           `yaml:"retry_attempts"`
    RetryDelay      time.Duration `yaml:"retry_delay"`
    BandwidthLimitUp   int        `yaml:"bandwidth_limit_up"`    // KB/s
    BandwidthLimitDown int        `yaml:"bandwidth_limit_down"`  // KB/s
}

type SyncFolderConfig struct {
    ID                 int      `yaml:"id"`
    Name               string   `yaml:"name"`
    LocalPath          string   `yaml:"local_path"`
    RemotePath         string   `yaml:"remote_path"`
    Direction          string   `yaml:"direction"` // bidirectional, upload_only, download_only
    Enabled            bool     `yaml:"enabled"`
    Excludes           []string `yaml:"excludes"`
    ConflictResolution string   `yaml:"conflict_resolution"`
    SyncMode           string   `yaml:"sync_mode"` // continuous, interval, scheduled
    SyncInterval       int      `yaml:"sync_interval"` // seconds
    SyncSchedule       string   `yaml:"sync_schedule"` // cron format
    BandwidthLimit     int      `yaml:"bandwidth_limit"` // KB/s
    MaxFileSize        int64    `yaml:"max_file_size"` // bytes
}

type NotificationSettings struct {
    Enabled      bool `yaml:"enabled"`
    ShowSuccess  bool `yaml:"show_success"`
    ShowErrors   bool `yaml:"show_errors"`
    ShowConflicts bool `yaml:"show_conflicts"`
}

type APIConfig struct {
    Endpoint string `yaml:"endpoint"`
    // API key loaded separately from main config
}
```

### Database Models

```go
// internal/db/models.go
package db

import "time"

type SyncFolder struct {
    ID                 int       `db:"id"`
    LocalPath          string    `db:"local_path"`
    RemotePath         string    `db:"remote_path"`
    Direction          string    `db:"direction"`
    Enabled            bool      `db:"enabled"`
    ConflictResolution string    `db:"conflict_resolution"`
    ExcludePatterns    string    `db:"exclude_patterns"` // JSON array
    BandwidthLimit     *int      `db:"bandwidth_limit"`  // nullable
    SyncInterval       *int      `db:"sync_interval"`    // nullable
    CreatedAt          time.Time `db:"created_at"`
    UpdatedAt          time.Time `db:"updated_at"`
}

type FileState struct {
    ID                int        `db:"id"`
    SyncFolderID      int        `db:"sync_folder_id"`
    RelativePath      string     `db:"relative_path"`
    LocalHash         *string    `db:"local_hash"`
    RemoteHash        *string    `db:"remote_hash"`
    LocalModifiedAt   *time.Time `db:"local_modified_at"`
    RemoteModifiedAt  *time.Time `db:"remote_modified_at"`
    LocalSize         *int64     `db:"local_size"`
    RemoteSize        *int64     `db:"remote_size"`
    SyncStatus        string     `db:"sync_status"` // synced, pending, conflict, error
    LastSyncedAt      *time.Time `db:"last_synced_at"`
    CreatedAt         time.Time  `db:"created_at"`
    UpdatedAt         time.Time  `db:"updated_at"`
}

type QueueOperation struct {
    ID            int        `db:"id"`
    SyncFolderID  int        `db:"sync_folder_id"`
    RelativePath  string     `db:"relative_path"`
    Operation     string     `db:"operation"` // upload, download, delete
    Priority      int        `db:"priority"`
    Attempts      int        `db:"attempts"`
    MaxAttempts   int        `db:"max_attempts"`
    Status        string     `db:"status"` // pending, processing, completed, failed
    ErrorMessage  *string    `db:"error_message"`
    CreatedAt     time.Time  `db:"created_at"`
    StartedAt     *time.Time `db:"started_at"`
    CompletedAt   *time.Time `db:"completed_at"`
}

type Activity struct {
    ID               int        `db:"id"`
    SyncFolderID     *int       `db:"sync_folder_id"`
    Operation        string     `db:"operation"`
    Path             string     `db:"path"`
    Status           string     `db:"status"` // success, error
    Details          *string    `db:"details"` // JSON
    ErrorMessage     *string    `db:"error_message"`
    BytesTransferred *int64     `db:"bytes_transferred"`
    DurationMS       *int       `db:"duration_ms"`
    CreatedAt        time.Time  `db:"created_at"`
}

type Conflict struct {
    ID                int        `db:"id"`
    SyncFolderID      int        `db:"sync_folder_id"`
    RelativePath      string     `db:"relative_path"`
    LocalHash         *string    `db:"local_hash"`
    RemoteHash        *string    `db:"remote_hash"`
    LocalModifiedAt   *time.Time `db:"local_modified_at"`
    RemoteModifiedAt  *time.Time `db:"remote_modified_at"`
    Resolution        *string    `db:"resolution"`
    Resolved          bool       `db:"resolved"`
    CreatedAt         time.Time  `db:"created_at"`
    ResolvedAt        *time.Time `db:"resolved_at"`
}
```

### Sync Engine Structures

```go
// internal/sync/types.go
package sync

import "time"

type FileEvent struct {
    Path      string
    EventType EventType
    Timestamp time.Time
    Hash      string
    Size      int64
}

type EventType int

const (
    EventCreate EventType = iota
    EventModify
    EventDelete
    EventRename
)

func (e EventType) String() string {
    return [...]string{"create", "modify", "delete", "rename"}[e]
}

type SyncAction int

const (
    ActionNone SyncAction = iota
    ActionUpload
    ActionDownload
    ActionDelete
    ActionConflict
)

func (a SyncAction) String() string {
    return [...]string{"none", "upload", "download", "delete", "conflict"}[a]
}

type SyncDirection int

const (
    DirectionBidirectional SyncDirection = iota
    DirectionUploadOnly
    DirectionDownloadOnly
)

func (d SyncDirection) String() string {
    return [...]string{"bidirectional", "upload_only", "download_only"}[d]
}

type ConflictResolution int

const (
    ResolutionKeepLocal ConflictResolution = iota
    ResolutionKeepRemote
    ResolutionKeepBoth
    ResolutionManual
)

func (r ConflictResolution) String() string {
    return [...]string{"keep_local", "keep_remote", "keep_both", "manual"}[r]
}

type OperationType int

const (
    OperationUpload OperationType = iota
    OperationDownload
    OperationDelete
)

func (o OperationType) String() string {
    return [...]string{"upload", "download", "delete"}[o]
}

type Operation struct {
    Type         OperationType
    LocalPath    string
    RemotePath   string
    SyncFolderID int
    Priority     int
}

type SyncResult struct {
    Success          bool
    BytesTransferred int64
    Duration         time.Duration
    Error            error
}

type ExcludeRules struct {
    patterns []string
    compiled []*regexp.Regexp
}
```

### IPC Protocol Structures

```go
// internal/ipc/protocol.go
package ipc

import "time"

type Command struct {
    Type string          `json:"command"`
    Data json.RawMessage `json:"data,omitempty"`
}

type Response struct {
    Success bool            `json:"success"`
    Data    json.RawMessage `json:"data,omitempty"`
    Error   string          `json:"error,omitempty"`
}

// Status command
type StatusRequest struct{}

type StatusResponse struct {
    DaemonRunning bool               `json:"daemon_running"`
    SyncFolders   []SyncFolderStatus `json:"sync_folders"`
    QueueSize     int                `json:"queue_size"`
    Uptime        time.Duration      `json:"uptime"`
}

type SyncFolderStatus struct {
    ID           int       `json:"id"`
    Name         string    `json:"name"`
    LocalPath    string    `json:"local_path"`
    RemotePath   string    `json:"remote_path"`
    Status       string    `json:"status"` // idle, syncing, paused, error
    FilesPending int       `json:"files_pending"`
    LastSync     time.Time `json:"last_sync"`
    ErrorMessage string    `json:"error_message,omitempty"`
}

// Add sync folder command
type AddSyncFolderRequest struct {
    LocalPath          string   `json:"local_path"`
    RemotePath         string   `json:"remote_path"`
    Direction          string   `json:"direction"`
    Excludes           []string `json:"excludes"`
    ConflictResolution string   `json:"conflict_resolution"`
    BandwidthLimit     int      `json:"bandwidth_limit,omitempty"`
}

type AddSyncFolderResponse struct {
    ID int `json:"id"`
}

// Remove sync folder command
type RemoveSyncFolderRequest struct {
    ID int `json:"id"`
}

type RemoveSyncFolderResponse struct {
    Success bool `json:"success"`
}

// Pause/Resume sync folder
type PauseSyncFolderRequest struct {
    ID int `json:"id"`
}

type ResumeSyncFolderRequest struct {
    ID int `json:"id"`
}

// Get activity log
type GetActivityRequest struct {
    Limit    int `json:"limit"`
    FolderID *int `json:"folder_id,omitempty"`
}

type GetActivityResponse struct {
    Activities []ActivityEntry `json:"activities"`
}

type ActivityEntry struct {
    ID               int       `json:"id"`
    Operation        string    `json:"operation"`
    Path             string    `json:"path"`
    Status           string    `json:"status"`
    BytesTransferred int64     `json:"bytes_transferred"`
    Timestamp        time.Time `json:"timestamp"`
    ErrorMessage     string    `json:"error_message,omitempty"`
}

// Get conflicts
type GetConflictsRequest struct{}

type GetConflictsResponse struct {
    Conflicts []ConflictEntry `json:"conflicts"`
}

type ConflictEntry struct {
    ID               int       `json:"id"`
    FolderID         int       `json:"folder_id"`
    Path             string    `json:"path"`
    LocalModified    time.Time `json:"local_modified"`
    RemoteModified   time.Time `json:"remote_modified"`
}

// Resolve conflict
type ResolveConflictRequest struct {
    ConflictID int    `json:"conflict_id"`
    Resolution string `json:"resolution"` // keep_local, keep_remote, keep_both
}

type ResolveConflictResponse struct {
    Success bool `json:"success"`
}

// Force sync
type ForceSyncRequest struct {
    FolderID int `json:"folder_id"`
}

type ForceSyncResponse struct {
    Success bool `json:"success"`
}
```

### API Client Structures

```go
// internal/api/types.go
package api

import "time"

type BucketInfo struct {
    Name      string    `json:"name"`
    Objects   int       `json:"objects"`
    Size      int64     `json:"size"`
    CreatedAt time.Time `json:"created_at"`
}

type FileInfo struct {
    Name         string    `json:"name"`
    Path         string    `json:"path"`
    Size         int64     `json:"size"`
    IsDir        bool      `json:"is_dir"`
    ModifiedAt   time.Time `json:"modified_at"`
    ContentType  string    `json:"content_type,omitempty"`
    ETag         string    `json:"etag,omitempty"`
}

type FileMetadata struct {
    Path         string            `json:"path"`
    Size         int64             `json:"size"`
    ModifiedAt   time.Time         `json:"modified_at"`
    ContentType  string            `json:"content_type"`
    ETag         string            `json:"etag"`
    Hash         string            `json:"hash"`
    CustomFields map[string]string `json:"custom_fields"`
}

type UploadOptions struct {
    ContentType    string
    Metadata       map[string]string
    ProgressFunc   func(bytesTransferred int64)
    BandwidthLimit int // KB/s
}

type DownloadOptions struct {
    ProgressFunc   func(bytesTransferred int64)
    BandwidthLimit int // KB/s
}

type ListOptions struct {
    Recursive bool
    Prefix    string
    MaxKeys   int
}
```

### GUI State Structures

```go
// cmd/gui/types.go
package main

type AppState struct {
    Connected       bool
    SyncFolders     []SyncFolderStatus
    RecentActivity  []ActivityEntry
    QueueSize       int
    StorageUsed     int64
    StorageLimit    int64
    CurrentView     string // dashboard, browser, sync, activity, settings
}

type NavigationItem struct {
    Label    string
    Icon     string
    ViewName string
}

type SyncFolderListItem struct {
    ID         int
    Name       string
    LocalPath  string
    RemotePath string
    Status     string
    Icon       string
    Color      color.Color
}
```

## Interface Definitions

```go
// internal/db/interfaces.go
package db

type Database interface {
    // Sync folders
    CreateSyncFolder(folder *SyncFolder) error
    GetSyncFolder(id int) (*SyncFolder, error)
    ListSyncFolders() ([]*SyncFolder, error)
    UpdateSyncFolder(folder *SyncFolder) error
    DeleteSyncFolder(id int) error

    // File states
    UpsertFileState(state *FileState) error
    GetFileState(folderID int, path string) (*FileState, error)
    ListFileStates(folderID int, status string) ([]*FileState, error)
    UpdateSyncStatus(id int, status string) error

    // Queue operations
    EnqueueOperation(op *QueueOperation) error
    DequeueOperation() (*QueueOperation, error)
    UpdateOperationStatus(id int, status string, errorMsg *string) error
    GetQueueSize() (int, error)

    // Activity
    LogActivity(activity *Activity) error
    GetRecentActivity(limit int) ([]*Activity, error)

    // Conflicts
    CreateConflict(conflict *Conflict) error
    GetUnresolvedConflicts() ([]*Conflict, error)
    ResolveConflict(id int, resolution string) error
}
```

```go
// internal/api/interfaces.go
package api

type StorageClient interface {
    // File operations
    UploadFile(localPath, remotePath string, opts *UploadOptions) error
    DownloadFile(remotePath, localPath string, opts *DownloadOptions) error
    DeleteFile(remotePath string) error
    CopyFile(srcPath, dstPath string) error
    MoveFile(srcPath, dstPath string) error

    // Metadata
    GetFileMetadata(remotePath string) (*FileMetadata, error)
    UpdateFileMetadata(remotePath string, metadata map[string]string) error

    // Listing
    ListBuckets() ([]*BucketInfo, error)
    ListFiles(remotePath string, opts *ListOptions) ([]*FileInfo, error)

    // Search
    SearchFiles(query string) ([]*FileInfo, error)
    SearchByHash(hash string) ([]*FileInfo, error)
}
```

```go
// internal/sync/interfaces.go
package sync

type SyncEngine interface {
    // Initialization
    Initialize(db Database, client StorageClient, config *Config) error
    Start() error
    Stop() error

    // Sync operations
    SyncFolder(folderID int) error
    ProcessFileEvent(event *FileEvent) error

    // File operations
    UploadFile(folderID int, relativePath string) (*SyncResult, error)
    DownloadFile(folderID int, relativePath string) (*SyncResult, error)
    DeleteFile(folderID int, relativePath string, remote bool) (*SyncResult, error)

    // Conflict handling
    DetectConflicts(folderID int) ([]*Conflict, error)
    ResolveConflict(conflict *Conflict, resolution ConflictResolution) error
}

type FileHasher interface {
    HashFile(path string) (string, error)
    HashFileChunked(path string, chunkSize int) (string, error)
}

type ExcludeChecker interface {
    ShouldExclude(path string) bool
    AddPattern(pattern string) error
    RemovePattern(pattern string) error
}
```

```go
// internal/ipc/interfaces.go
package ipc

type Server interface {
    Start() error
    Stop() error
    RegisterHandler(command string, handler HandlerFunc)
}

type Client interface {
    Connect() error
    SendCommand(cmd *Command) (*Response, error)
    Close() error
}

type HandlerFunc func(data json.RawMessage) (*Response, error)
```

## Constants

```go
// internal/sync/constants.go
package sync

const (
    // Sync status values
    StatusSynced   = "synced"
    StatusPending  = "pending"
    StatusConflict = "conflict"
    StatusError    = "error"

    // Queue status values
    QueuePending    = "pending"
    QueueProcessing = "processing"
    QueueCompleted  = "completed"
    QueueFailed     = "failed"

    // Activity status values
    ActivitySuccess = "success"
    ActivityError   = "error"

    // Default values
    DefaultDebounceDelay = 3 * time.Second
    DefaultWorkerCount   = 4
    DefaultMaxRetries    = 3
    DefaultRetryDelay    = 5 * time.Second
    DefaultChunkSize     = 1024 * 1024 // 1MB
)
```

## Error Types

```go
// internal/errors/errors.go
package errors

import "errors"

var (
    ErrNotFound          = errors.New("not found")
    ErrAlreadyExists     = errors.New("already exists")
    ErrInvalidPath       = errors.New("invalid path")
    ErrConflict          = errors.New("sync conflict")
    ErrDatabaseError     = errors.New("database error")
    ErrAPIError          = errors.New("api error")
    ErrNetworkError      = errors.New("network error")
    ErrPermissionDenied  = errors.New("permission denied")
    ErrDaemonNotRunning  = errors.New("daemon not running")
    ErrInvalidConfig     = errors.New("invalid configuration")
)

type SyncError struct {
    Operation string
    Path      string
    Err       error
}

func (e *SyncError) Error() string {
    return fmt.Sprintf("%s failed for %s: %v", e.Operation, e.Path, e.Err)
}

func (e *SyncError) Unwrap() error {
    return e.Err
}
```

## Usage Examples

```go
// Example: Creating a sync folder
folder := &db.SyncFolder{
    LocalPath:          "/Users/ryan/Documents",
    RemotePath:         "my-bucket/Documents",
    Direction:          sync.DirectionBidirectional.String(),
    Enabled:            true,
    ConflictResolution: sync.ResolutionKeepLocal.String(),
    ExcludePatterns:    `["*.tmp", ".DS_Store"]`,
}
err := database.CreateSyncFolder(folder)

// Example: Sending IPC command
client, _ := ipc.NewClient("/path/to/socket")
cmd := &ipc.Command{
    Type: "status",
}
response, err := client.SendCommand(cmd)

// Example: Processing file event
event := &sync.FileEvent{
    Path:      "/Users/ryan/Documents/file.txt",
    EventType: sync.EventModify,
    Timestamp: time.Now(),
}
err := syncEngine.ProcessFileEvent(event)
```
