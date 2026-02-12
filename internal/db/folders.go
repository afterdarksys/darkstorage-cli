package db

import (
	"database/sql"
	"time"
)

func (db *DB) CreateSyncFolder(folder *SyncFolder) error {
	result, err := db.conn.Exec(`
		INSERT INTO sync_folders (
			local_path, remote_path, direction, enabled,
			conflict_resolution, exclude_patterns, bandwidth_limit, sync_interval
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, folder.LocalPath, folder.RemotePath, folder.Direction, folder.Enabled,
		folder.ConflictResolution, folder.ExcludePatterns, folder.BandwidthLimit, folder.SyncInterval)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	folder.ID = int(id)
	return nil
}

func (db *DB) GetSyncFolder(id int) (*SyncFolder, error) {
	folder := &SyncFolder{}
	err := db.conn.QueryRow(`
		SELECT id, local_path, remote_path, direction, enabled,
			conflict_resolution, exclude_patterns, bandwidth_limit, sync_interval,
			created_at, updated_at
		FROM sync_folders WHERE id = ?
	`, id).Scan(
		&folder.ID, &folder.LocalPath, &folder.RemotePath, &folder.Direction, &folder.Enabled,
		&folder.ConflictResolution, &folder.ExcludePatterns, &folder.BandwidthLimit, &folder.SyncInterval,
		&folder.CreatedAt, &folder.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return folder, err
}

func (db *DB) ListSyncFolders() ([]*SyncFolder, error) {
	rows, err := db.conn.Query(`
		SELECT id, local_path, remote_path, direction, enabled,
			conflict_resolution, exclude_patterns, bandwidth_limit, sync_interval,
			created_at, updated_at
		FROM sync_folders ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []*SyncFolder
	for rows.Next() {
		folder := &SyncFolder{}
		err := rows.Scan(
			&folder.ID, &folder.LocalPath, &folder.RemotePath, &folder.Direction, &folder.Enabled,
			&folder.ConflictResolution, &folder.ExcludePatterns, &folder.BandwidthLimit, &folder.SyncInterval,
			&folder.CreatedAt, &folder.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		folders = append(folders, folder)
	}
	return folders, rows.Err()
}

func (db *DB) UpdateSyncFolder(folder *SyncFolder) error {
	_, err := db.conn.Exec(`
		UPDATE sync_folders SET
			local_path = ?, remote_path = ?, direction = ?, enabled = ?,
			conflict_resolution = ?, exclude_patterns = ?, bandwidth_limit = ?,
			sync_interval = ?, updated_at = ?
		WHERE id = ?
	`, folder.LocalPath, folder.RemotePath, folder.Direction, folder.Enabled,
		folder.ConflictResolution, folder.ExcludePatterns, folder.BandwidthLimit,
		folder.SyncInterval, time.Now(), folder.ID)
	return err
}

func (db *DB) DeleteSyncFolder(id int) error {
	_, err := db.conn.Exec("DELETE FROM sync_folders WHERE id = ?", id)
	return err
}
