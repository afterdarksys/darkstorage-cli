package ipc

import (
	"encoding/json"
	"time"
)

type Command struct {
	Type string          `json:"command"`
	Data json.RawMessage `json:"data,omitempty"`
}

type Response struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type StatusRequest struct{}

type StatusResponse struct {
	DaemonRunning bool               `json:"daemon_running"`
	SyncFolders   []SyncFolderStatus `json:"sync_folders"`
	QueueSize     int                `json:"queue_size"`
	Uptime        string             `json:"uptime"`
}

type SyncFolderStatus struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	LocalPath    string    `json:"local_path"`
	RemotePath   string    `json:"remote_path"`
	Status       string    `json:"status"`
	FilesPending int       `json:"files_pending"`
	LastSync     time.Time `json:"last_sync"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

type AddSyncFolderRequest struct {
	Name               string   `json:"name"`
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

type RemoveSyncFolderRequest struct {
	ID int `json:"id"`
}

type GetConfigRequest struct{}

type GetConfigResponse struct {
	Config map[string]interface{} `json:"config"`
}

type SetConfigRequest struct {
	Config map[string]interface{} `json:"config"`
}

type SetConfigResponse struct {
	Success bool `json:"success"`
}

type GetActivityRequest struct {
	Limit    int `json:"limit"`
	FolderID *int `json:"folder_id,omitempty"`
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

type GetActivityResponse struct {
	Activities []ActivityEntry `json:"activities"`
}

type ForceSyncRequest struct {
	FolderID int `json:"folder_id"`
}
