package db

import "time"

type SyncFolder struct {
	ID                 int       `db:"id"`
	LocalPath          string    `db:"local_path"`
	RemotePath         string    `db:"remote_path"`
	Direction          string    `db:"direction"`
	Enabled            bool      `db:"enabled"`
	ConflictResolution string    `db:"conflict_resolution"`
	ExcludePatterns    string    `db:"exclude_patterns"`
	BandwidthLimit     *int      `db:"bandwidth_limit"`
	SyncInterval       *int      `db:"sync_interval"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

type FileState struct {
	ID               int        `db:"id"`
	SyncFolderID     int        `db:"sync_folder_id"`
	RelativePath     string     `db:"relative_path"`
	LocalHash        *string    `db:"local_hash"`
	RemoteHash       *string    `db:"remote_hash"`
	LocalModifiedAt  *time.Time `db:"local_modified_at"`
	RemoteModifiedAt *time.Time `db:"remote_modified_at"`
	LocalSize        *int64     `db:"local_size"`
	RemoteSize       *int64     `db:"remote_size"`
	SyncStatus       string     `db:"sync_status"`
	LastSyncedAt     *time.Time `db:"last_synced_at"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
}

type QueueOperation struct {
	ID           int        `db:"id"`
	SyncFolderID int        `db:"sync_folder_id"`
	RelativePath string     `db:"relative_path"`
	Operation    string     `db:"operation"`
	Priority     int        `db:"priority"`
	Attempts     int        `db:"attempts"`
	MaxAttempts  int        `db:"max_attempts"`
	Status       string     `db:"status"`
	ErrorMessage *string    `db:"error_message"`
	CreatedAt    time.Time  `db:"created_at"`
	StartedAt    *time.Time `db:"started_at"`
	CompletedAt  *time.Time `db:"completed_at"`
}

type Activity struct {
	ID               int        `db:"id"`
	SyncFolderID     *int       `db:"sync_folder_id"`
	Operation        string     `db:"operation"`
	Path             string     `db:"path"`
	Status           string     `db:"status"`
	Details          *string    `db:"details"`
	ErrorMessage     *string    `db:"error_message"`
	BytesTransferred *int64     `db:"bytes_transferred"`
	DurationMS       *int       `db:"duration_ms"`
	CreatedAt        time.Time  `db:"created_at"`
}

type Conflict struct {
	ID               int        `db:"id"`
	SyncFolderID     int        `db:"sync_folder_id"`
	RelativePath     string     `db:"relative_path"`
	LocalHash        *string    `db:"local_hash"`
	RemoteHash       *string    `db:"remote_hash"`
	LocalModifiedAt  *time.Time `db:"local_modified_at"`
	RemoteModifiedAt *time.Time `db:"remote_modified_at"`
	Resolution       *string    `db:"resolution"`
	Resolved         bool       `db:"resolved"`
	CreatedAt        time.Time  `db:"created_at"`
	ResolvedAt       *time.Time `db:"resolved_at"`
}
