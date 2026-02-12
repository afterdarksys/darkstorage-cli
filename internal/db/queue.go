package db

import (
	"database/sql"
	"time"
)

func (db *DB) EnqueueOperation(op *QueueOperation) error {
	result, err := db.conn.Exec(`
		INSERT INTO sync_queue (
			sync_folder_id, relative_path, operation, priority, max_attempts
		) VALUES (?, ?, ?, ?, ?)
	`, op.SyncFolderID, op.RelativePath, op.Operation, op.Priority, op.MaxAttempts)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	op.ID = int(id)
	return nil
}

func (db *DB) DequeueOperation() (*QueueOperation, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	op := &QueueOperation{}
	err = tx.QueryRow(`
		SELECT id, sync_folder_id, relative_path, operation, priority,
			attempts, max_attempts, status, error_message, created_at, started_at, completed_at
		FROM sync_queue
		WHERE status = 'pending' AND attempts < max_attempts
		ORDER BY priority DESC, created_at ASC
		LIMIT 1
	`).Scan(
		&op.ID, &op.SyncFolderID, &op.RelativePath, &op.Operation, &op.Priority,
		&op.Attempts, &op.MaxAttempts, &op.Status, &op.ErrorMessage,
		&op.CreatedAt, &op.StartedAt, &op.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	now := time.Now()
	_, err = tx.Exec(`
		UPDATE sync_queue SET status = 'processing', started_at = ?, attempts = attempts + 1
		WHERE id = ?
	`, now, op.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	op.Status = "processing"
	op.StartedAt = &now
	op.Attempts++
	return op, nil
}

func (db *DB) UpdateOperationStatus(id int, status string, errorMsg *string) error {
	now := time.Now()
	_, err := db.conn.Exec(`
		UPDATE sync_queue SET status = ?, error_message = ?, completed_at = ?
		WHERE id = ?
	`, status, errorMsg, now, id)
	return err
}

func (db *DB) GetQueueSize() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM sync_queue WHERE status = 'pending'").Scan(&count)
	return count, err
}

func (db *DB) ClearCompletedOperations(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	_, err := db.conn.Exec(`
		DELETE FROM sync_queue WHERE status = 'completed' AND completed_at < ?
	`, cutoff)
	return err
}
