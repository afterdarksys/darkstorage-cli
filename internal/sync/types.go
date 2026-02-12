package sync

import "time"

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

type FileEvent struct {
	Path      string
	EventType EventType
	Timestamp time.Time
	Hash      string
	Size      int64
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

const (
	StatusSynced   = "synced"
	StatusPending  = "pending"
	StatusConflict = "conflict"
	StatusError    = "error"

	QueuePending    = "pending"
	QueueProcessing = "processing"
	QueueCompleted  = "completed"
	QueueFailed     = "failed"

	DefaultDebounceDelay = 3 * time.Second
	DefaultWorkerCount   = 4
)
