package db

func (db *DB) migrate() error {
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	var version int
	err = db.conn.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
	if err != nil {
		return err
	}

	migrations := []string{
		// Version 1: sync_folders table
		`CREATE TABLE sync_folders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			local_path TEXT NOT NULL UNIQUE,
			remote_path TEXT NOT NULL,
			direction TEXT NOT NULL,
			enabled INTEGER DEFAULT 1,
			conflict_resolution TEXT DEFAULT 'keep_local',
			exclude_patterns TEXT,
			bandwidth_limit INTEGER,
			sync_interval INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Version 2: file_states table
		`CREATE TABLE file_states (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sync_folder_id INTEGER NOT NULL,
			relative_path TEXT NOT NULL,
			local_hash TEXT,
			remote_hash TEXT,
			local_modified_at DATETIME,
			remote_modified_at DATETIME,
			local_size INTEGER,
			remote_size INTEGER,
			sync_status TEXT DEFAULT 'pending',
			last_synced_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (sync_folder_id) REFERENCES sync_folders(id) ON DELETE CASCADE,
			UNIQUE(sync_folder_id, relative_path)
		)`,
		// Version 3: sync_queue table
		`CREATE TABLE sync_queue (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sync_folder_id INTEGER NOT NULL,
			relative_path TEXT NOT NULL,
			operation TEXT NOT NULL,
			priority INTEGER DEFAULT 0,
			attempts INTEGER DEFAULT 0,
			max_attempts INTEGER DEFAULT 3,
			status TEXT DEFAULT 'pending',
			error_message TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			started_at DATETIME,
			completed_at DATETIME,
			FOREIGN KEY (sync_folder_id) REFERENCES sync_folders(id) ON DELETE CASCADE
		)`,
		// Version 4: activity_log table
		`CREATE TABLE activity_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sync_folder_id INTEGER,
			operation TEXT NOT NULL,
			path TEXT NOT NULL,
			status TEXT NOT NULL,
			details TEXT,
			error_message TEXT,
			bytes_transferred INTEGER,
			duration_ms INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (sync_folder_id) REFERENCES sync_folders(id) ON DELETE SET NULL
		)`,
		// Version 5: conflicts table
		`CREATE TABLE conflicts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sync_folder_id INTEGER NOT NULL,
			relative_path TEXT NOT NULL,
			local_hash TEXT,
			remote_hash TEXT,
			local_modified_at DATETIME,
			remote_modified_at DATETIME,
			resolution TEXT,
			resolved INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			resolved_at DATETIME,
			FOREIGN KEY (sync_folder_id) REFERENCES sync_folders(id) ON DELETE CASCADE
		)`,
		// Version 6: Indexes
		`CREATE INDEX idx_file_states_folder ON file_states(sync_folder_id)`,
		`CREATE INDEX idx_file_states_status ON file_states(sync_status)`,
		`CREATE INDEX idx_sync_queue_status ON sync_queue(status)`,
		`CREATE INDEX idx_activity_log_created ON activity_log(created_at DESC)`,
	}

	for i := version; i < len(migrations); i++ {
		if _, err := db.conn.Exec(migrations[i]); err != nil {
			return err
		}
		if _, err := db.conn.Exec("INSERT INTO schema_version (version) VALUES (?)", i+1); err != nil {
			return err
		}
	}

	return nil
}
