package db

import (
	"database/sql"
	"time"
)

func (db *DB) CreateConflict(conflict *Conflict) error {
	result, err := db.conn.Exec(`
		INSERT INTO conflicts (
			sync_folder_id, relative_path, local_hash, remote_hash,
			local_modified_at, remote_modified_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`, conflict.SyncFolderID, conflict.RelativePath, conflict.LocalHash, conflict.RemoteHash,
		conflict.LocalModifiedAt, conflict.RemoteModifiedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	conflict.ID = int(id)
	return nil
}

func (db *DB) GetUnresolvedConflicts() ([]*Conflict, error) {
	rows, err := db.conn.Query(`
		SELECT id, sync_folder_id, relative_path, local_hash, remote_hash,
			local_modified_at, remote_modified_at, resolution, resolved,
			created_at, resolved_at
		FROM conflicts
		WHERE resolved = 0
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conflicts []*Conflict
	for rows.Next() {
		conflict := &Conflict{}
		err := rows.Scan(
			&conflict.ID, &conflict.SyncFolderID, &conflict.RelativePath,
			&conflict.LocalHash, &conflict.RemoteHash,
			&conflict.LocalModifiedAt, &conflict.RemoteModifiedAt,
			&conflict.Resolution, &conflict.Resolved,
			&conflict.CreatedAt, &conflict.ResolvedAt,
		)
		if err != nil {
			return nil, err
		}
		conflicts = append(conflicts, conflict)
	}
	return conflicts, rows.Err()
}

func (db *DB) ResolveConflict(id int, resolution string) error {
	now := time.Now()
	_, err := db.conn.Exec(`
		UPDATE conflicts SET resolution = ?, resolved = 1, resolved_at = ?
		WHERE id = ?
	`, resolution, now, id)
	return err
}

func (db *DB) GetConflict(id int) (*Conflict, error) {
	conflict := &Conflict{}
	err := db.conn.QueryRow(`
		SELECT id, sync_folder_id, relative_path, local_hash, remote_hash,
			local_modified_at, remote_modified_at, resolution, resolved,
			created_at, resolved_at
		FROM conflicts WHERE id = ?
	`, id).Scan(
		&conflict.ID, &conflict.SyncFolderID, &conflict.RelativePath,
		&conflict.LocalHash, &conflict.RemoteHash,
		&conflict.LocalModifiedAt, &conflict.RemoteModifiedAt,
		&conflict.Resolution, &conflict.Resolved,
		&conflict.CreatedAt, &conflict.ResolvedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return conflict, err
}
