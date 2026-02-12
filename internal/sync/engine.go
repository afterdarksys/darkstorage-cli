package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/darkstorage/cli/internal/api"
	"github.com/darkstorage/cli/internal/db"
)

type Engine struct {
	db     *db.DB
	client *api.Client
}

func NewEngine(database *db.DB, client *api.Client) *Engine {
	return &Engine{
		db:     database,
		client: client,
	}
}

func (e *Engine) SyncFolder(folderID int) error {
	folder, err := e.db.GetSyncFolder(folderID)
	if err != nil {
		return err
	}
	if folder == nil {
		return fmt.Errorf("folder not found: %d", folderID)
	}

	fmt.Printf("Syncing folder: %s -> %s\n", folder.LocalPath, folder.RemotePath)

	return e.scanAndSync(folder)
}

func (e *Engine) scanAndSync(folder *db.SyncFolder) error {
	return filepath.Walk(folder.LocalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(folder.LocalPath, path)
		if err != nil {
			return err
		}

		hash, err := HashFile(path)
		if err != nil {
			return err
		}

		modTime := info.ModTime()
		size := info.Size()

		state := &db.FileState{
			SyncFolderID:    folder.ID,
			RelativePath:    relPath,
			LocalHash:       &hash,
			LocalModifiedAt: &modTime,
			LocalSize:       &size,
			SyncStatus:      StatusPending,
		}

		if err := e.db.UpsertFileState(state); err != nil {
			return err
		}

		op := &db.QueueOperation{
			SyncFolderID: folder.ID,
			RelativePath: relPath,
			Operation:    "upload",
			Priority:     0,
			MaxAttempts:  3,
		}
		return e.db.EnqueueOperation(op)
	})
}

func (e *Engine) ProcessFileEvent(event *FileEvent, folderID int) error {
	fmt.Printf("Processing event: %s %s\n", event.EventType, event.Path)

	op := &db.QueueOperation{
		SyncFolderID: folderID,
		RelativePath: event.Path,
		Operation:    "upload",
		Priority:     0,
		MaxAttempts:  3,
	}

	return e.db.EnqueueOperation(op)
}

func (e *Engine) ProcessQueue() error {
	for {
		op, err := e.db.DequeueOperation()
		if err != nil {
			return err
		}
		if op == nil {
			break
		}

		startTime := time.Now()
		err = e.executeOperation(op)
		duration := time.Since(startTime)

		activity := &db.Activity{
			SyncFolderID: &op.SyncFolderID,
			Operation:    op.Operation,
			Path:         op.RelativePath,
			DurationMS:   intPtr(int(duration.Milliseconds())),
		}

		if err != nil {
			errMsg := err.Error()
			activity.Status = "error"
			activity.ErrorMessage = &errMsg
			e.db.UpdateOperationStatus(op.ID, QueueFailed, &errMsg)
		} else {
			activity.Status = "success"
			e.db.UpdateOperationStatus(op.ID, QueueCompleted, nil)
		}

		e.db.LogActivity(activity)
	}
	return nil
}

func (e *Engine) executeOperation(op *db.QueueOperation) error {
	folder, err := e.db.GetSyncFolder(op.SyncFolderID)
	if err != nil {
		return err
	}

	localPath := filepath.Join(folder.LocalPath, op.RelativePath)
	remotePath := filepath.Join(folder.RemotePath, op.RelativePath)

	switch op.Operation {
	case "upload":
		return e.client.UploadFile(localPath, remotePath, nil)
	case "download":
		return e.client.DownloadFile(remotePath, localPath, nil)
	case "delete":
		return e.client.DeleteFile(remotePath)
	default:
		return fmt.Errorf("unknown operation: %s", op.Operation)
	}
}

func intPtr(i int) *int {
	return &i
}
