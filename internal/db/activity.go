package db

func (db *DB) LogActivity(activity *Activity) error {
	_, err := db.conn.Exec(`
		INSERT INTO activity_log (
			sync_folder_id, operation, path, status, details,
			error_message, bytes_transferred, duration_ms
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, activity.SyncFolderID, activity.Operation, activity.Path, activity.Status,
		activity.Details, activity.ErrorMessage, activity.BytesTransferred, activity.DurationMS)
	return err
}

func (db *DB) GetRecentActivity(limit int) ([]*Activity, error) {
	rows, err := db.conn.Query(`
		SELECT id, sync_folder_id, operation, path, status, details,
			error_message, bytes_transferred, duration_ms, created_at
		FROM activity_log
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*Activity
	for rows.Next() {
		activity := &Activity{}
		err := rows.Scan(
			&activity.ID, &activity.SyncFolderID, &activity.Operation, &activity.Path,
			&activity.Status, &activity.Details, &activity.ErrorMessage,
			&activity.BytesTransferred, &activity.DurationMS, &activity.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}
	return activities, rows.Err()
}

func (db *DB) GetActivityByFolder(folderID int, limit int) ([]*Activity, error) {
	rows, err := db.conn.Query(`
		SELECT id, sync_folder_id, operation, path, status, details,
			error_message, bytes_transferred, duration_ms, created_at
		FROM activity_log
		WHERE sync_folder_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, folderID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*Activity
	for rows.Next() {
		activity := &Activity{}
		err := rows.Scan(
			&activity.ID, &activity.SyncFolderID, &activity.Operation, &activity.Path,
			&activity.Status, &activity.Details, &activity.ErrorMessage,
			&activity.BytesTransferred, &activity.DurationMS, &activity.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}
	return activities, rows.Err()
}
