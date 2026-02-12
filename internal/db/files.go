package db

import (
	"database/sql"
	"time"
)

func (db *DB) UpsertFileState(state *FileState) error {
	state.UpdatedAt = time.Now()
	_, err := db.conn.Exec(`
		INSERT INTO file_states (
			sync_folder_id, relative_path, local_hash, remote_hash,
			local_modified_at, remote_modified_at, local_size, remote_size,
			sync_status, last_synced_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(sync_folder_id, relative_path) DO UPDATE SET
			local_hash = excluded.local_hash,
			remote_hash = excluded.remote_hash,
			local_modified_at = excluded.local_modified_at,
			remote_modified_at = excluded.remote_modified_at,
			local_size = excluded.local_size,
			remote_size = excluded.remote_size,
			sync_status = excluded.sync_status,
			last_synced_at = excluded.last_synced_at,
			updated_at = excluded.updated_at
	`, state.SyncFolderID, state.RelativePath, state.LocalHash, state.RemoteHash,
		state.LocalModifiedAt, state.RemoteModifiedAt, state.LocalSize, state.RemoteSize,
		state.SyncStatus, state.LastSyncedAt, state.UpdatedAt)
	return err
}

func (db *DB) GetFileState(folderID int, path string) (*FileState, error) {
	state := &FileState{}
	err := db.conn.QueryRow(`
		SELECT id, sync_folder_id, relative_path, local_hash, remote_hash,
			local_modified_at, remote_modified_at, local_size, remote_size,
			sync_status, last_synced_at, created_at, updated_at
		FROM file_states WHERE sync_folder_id = ? AND relative_path = ?
	`, folderID, path).Scan(
		&state.ID, &state.SyncFolderID, &state.RelativePath, &state.LocalHash, &state.RemoteHash,
		&state.LocalModifiedAt, &state.RemoteModifiedAt, &state.LocalSize, &state.RemoteSize,
		&state.SyncStatus, &state.LastSyncedAt, &state.CreatedAt, &state.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return state, err
}

func (db *DB) ListFileStates(folderID int, status string) ([]*FileState, error) {
	query := `
		SELECT id, sync_folder_id, relative_path, local_hash, remote_hash,
			local_modified_at, remote_modified_at, local_size, remote_size,
			sync_status, last_synced_at, created_at, updated_at
		FROM file_states WHERE sync_folder_id = ?
	`
	args := []interface{}{folderID}
	if status != "" {
		query += " AND sync_status = ?"
		args = append(args, status)
	}
	query += " ORDER BY relative_path"

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []*FileState
	for rows.Next() {
		state := &FileState{}
		err := rows.Scan(
			&state.ID, &state.SyncFolderID, &state.RelativePath, &state.LocalHash, &state.RemoteHash,
			&state.LocalModifiedAt, &state.RemoteModifiedAt, &state.LocalSize, &state.RemoteSize,
			&state.SyncStatus, &state.LastSyncedAt, &state.CreatedAt, &state.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		states = append(states, state)
	}
	return states, rows.Err()
}

func (db *DB) UpdateSyncStatus(id int, status string) error {
	now := time.Now()
	_, err := db.conn.Exec(`
		UPDATE file_states SET sync_status = ?, last_synced_at = ?, updated_at = ?
		WHERE id = ?
	`, status, now, now, id)
	return err
}
