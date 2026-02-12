# Dark Storage GUI & Daemon Design

## Overview

Dark Storage GUI is a desktop application built with Fyne/Go that provides:
- Visual interface for Dark Storage operations
- Background daemon for automatic directory synchronization
- SQLite database for tracking sync state and file history
- Configuration-driven sync rules

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                  Fyne GUI Application                │
│  ┌──────────────┬──────────────┬─────────────────┐ │
│  │  Dashboard   │  File Browser│  Sync Settings  │ │
│  │  View        │  View        │  View           │ │
│  └──────────────┴──────────────┴─────────────────┘ │
└────────────────┬────────────────────────────────────┘
                 │ IPC (Unix Socket / Named Pipe)
┌────────────────┴────────────────────────────────────┐
│              Daemon Service                          │
│  ┌──────────────────┬──────────────────────────┐   │
│  │  Sync Engine     │  File Watcher (fsnotify) │   │
│  ├──────────────────┼──────────────────────────┤   │
│  │  Config Manager  │  State Manager           │   │
│  └──────────────────┴──────────────────────────┘   │
└────────────────┬────────────────────────────────────┘
                 │
┌────────────────┴────────────────────────────────────┐
│            Storage Layer                             │
│  ┌──────────────────┬──────────────────────────┐   │
│  │  SQLite DB       │  Dark Storage API Client │   │
│  │  (local state)   │  (S3-compatible)         │   │
│  └──────────────────┴──────────────────────────┘   │
└─────────────────────────────────────────────────────┘
```

## Component Details

### 1. Fyne GUI Application

#### Main Window Structure
```
┌─────────────────────────────────────────────────┐
│  Dark Storage                        ⚙ [user]   │
├───────────┬─────────────────────────────────────┤
│           │                                      │
│ Dashboard │  Storage Usage: 45.2 GB / 100 GB   │
│           │  ┌───────────────────────────────┐  │
│ Files     │  │ [==========>          ] 45%   │  │
│           │  └───────────────────────────────┘  │
│ Sync      │                                      │
│           │  Recent Activity:                    │
│ Settings  │  • Uploaded: project.zip (2.3 MB)   │
│           │  • Synced: Documents/ (12 files)    │
│ Activity  │  • Deleted: old_backup.tar.gz       │
│           │                                      │
│ About     │  Sync Status:                        │
│           │  ✓ Documents: Up to date             │
│           │  ↻ Projects: Syncing (3/10 files)   │
│           │  ✓ Photos: Up to date                │
└───────────┴─────────────────────────────────────┘
```

#### Views

**Dashboard View**
- Storage usage meter
- Recent activity feed
- Quick stats (files synced, bandwidth used, errors)
- Sync status for all watched folders

**File Browser View**
- Tree view of buckets and folders
- File list with details (size, modified date, sync status)
- Context menu: upload, download, delete, share, view properties
- Drag & drop support for uploads
- Search and filter

**Sync Settings View**
- List of watched directories
- Add/remove sync folders
- Configure sync rules per folder:
  - Sync direction (bi-directional, upload-only, download-only)
  - Exclude patterns (*.tmp, .DS_Store, etc.)
  - Conflict resolution (keep local, keep remote, rename)
  - Bandwidth limits
  - Sync schedule (continuous, interval, scheduled)

**Activity Log View**
- Detailed log of all operations
- Filter by type (upload, download, delete, error)
- Export logs

**Settings View**
- API credentials
- Endpoint configuration
- Daemon control (start, stop, restart)
- Notification preferences
- Storage location for cache

### 2. Daemon Service

#### Core Responsibilities
- Monitor configured directories using fsnotify
- Detect file changes (create, modify, delete, rename)
- Queue sync operations
- Execute uploads/downloads
- Handle conflicts
- Maintain sync state in SQLite
- Provide IPC interface for GUI

#### Sync Engine Logic

**File Change Detection**
```go
type FileEvent struct {
    Path      string
    EventType string // create, modify, delete, rename
    Timestamp time.Time
    Hash      string
    Size      int64
}
```

**Sync Strategy**
1. Debounce rapid file changes (wait 3s after last change)
2. Calculate file hash (SHA-256)
3. Check database for last known state
4. Determine action needed:
   - New file → Upload
   - Modified (hash changed) → Upload
   - Deleted → Delete remote or trash
   - Conflict (both changed) → Apply conflict resolution

**Conflict Resolution**
- Keep Local: Overwrite remote with local
- Keep Remote: Download remote, overwrite local
- Keep Both: Rename with timestamp suffix
- Manual: Pause sync, notify user via GUI

**Queue Management**
- Priority queue: deletes > uploads > downloads
- Retry logic with exponential backoff
- Rate limiting per sync folder
- Parallel uploads (configurable worker count)

#### IPC Protocol

**Unix Socket** (Linux/macOS) or **Named Pipe** (Windows)

Commands:
```json
// Status request
{"command": "status"}

// Response
{
    "daemon_running": true,
    "sync_folders": [
        {
            "id": 1,
            "local_path": "/Users/ryan/Documents",
            "remote_path": "my-bucket/Documents",
            "status": "syncing",
            "files_pending": 3,
            "last_sync": "2024-01-05T10:30:00Z"
        }
    ],
    "queue_size": 5
}

// Add sync folder
{
    "command": "add_sync_folder",
    "local_path": "/path/to/folder",
    "remote_path": "bucket/path",
    "config": {
        "direction": "bidirectional",
        "excludes": ["*.tmp", ".DS_Store"],
        "conflict_resolution": "keep_local"
    }
}

// Pause/resume sync
{"command": "pause_sync", "folder_id": 1}
{"command": "resume_sync", "folder_id": 1}
```

### 3. SQLite Database Schema

```sql
-- Sync folder configurations
CREATE TABLE sync_folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    local_path TEXT NOT NULL UNIQUE,
    remote_path TEXT NOT NULL,
    direction TEXT NOT NULL, -- bidirectional, upload, download
    enabled INTEGER DEFAULT 1,
    conflict_resolution TEXT DEFAULT 'keep_local',
    exclude_patterns TEXT, -- JSON array
    bandwidth_limit INTEGER, -- KB/s, NULL = unlimited
    sync_interval INTEGER, -- seconds, NULL = continuous
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- File sync state
CREATE TABLE file_states (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sync_folder_id INTEGER NOT NULL,
    relative_path TEXT NOT NULL,
    local_hash TEXT,
    remote_hash TEXT,
    local_modified_at DATETIME,
    remote_modified_at DATETIME,
    local_size INTEGER,
    remote_size INTEGER,
    sync_status TEXT, -- synced, pending, conflict, error
    last_synced_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sync_folder_id) REFERENCES sync_folders(id) ON DELETE CASCADE,
    UNIQUE(sync_folder_id, relative_path)
);

-- Sync operation queue
CREATE TABLE sync_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sync_folder_id INTEGER NOT NULL,
    relative_path TEXT NOT NULL,
    operation TEXT NOT NULL, -- upload, download, delete
    priority INTEGER DEFAULT 0,
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    status TEXT DEFAULT 'pending', -- pending, processing, completed, failed
    error_message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    started_at DATETIME,
    completed_at DATETIME,
    FOREIGN KEY (sync_folder_id) REFERENCES sync_folders(id) ON DELETE CASCADE
);

-- Activity log
CREATE TABLE activity_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sync_folder_id INTEGER,
    operation TEXT NOT NULL,
    path TEXT NOT NULL,
    status TEXT NOT NULL, -- success, error
    details TEXT, -- JSON with additional info
    error_message TEXT,
    bytes_transferred INTEGER,
    duration_ms INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sync_folder_id) REFERENCES sync_folders(id) ON DELETE SET NULL
);

-- Conflict records
CREATE TABLE conflicts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sync_folder_id INTEGER NOT NULL,
    relative_path TEXT NOT NULL,
    local_hash TEXT,
    remote_hash TEXT,
    local_modified_at DATETIME,
    remote_modified_at DATETIME,
    resolution TEXT, -- manual, auto_local, auto_remote, keep_both
    resolved INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    resolved_at DATETIME,
    FOREIGN KEY (sync_folder_id) REFERENCES sync_folders(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX idx_file_states_folder ON file_states(sync_folder_id);
CREATE INDEX idx_file_states_status ON file_states(sync_status);
CREATE INDEX idx_sync_queue_status ON sync_queue(status);
CREATE INDEX idx_activity_log_created ON activity_log(created_at DESC);
```

### 4. Configuration File

**Location**: `~/.darkstorage/daemon.yaml`

```yaml
daemon:
  # Daemon settings
  enabled: true
  log_level: info
  log_file: ~/.darkstorage/daemon.log
  pid_file: ~/.darkstorage/daemon.pid
  ipc_socket: ~/.darkstorage/daemon.sock

  # Performance settings
  worker_threads: 4
  max_queue_size: 1000
  debounce_delay: 3s

  # Network settings
  timeout: 30s
  retry_attempts: 3
  retry_delay: 5s

  # Global bandwidth limits (optional)
  bandwidth_limit_up: 1024 # KB/s
  bandwidth_limit_down: 2048 # KB/s

# Sync folder definitions (also managed via GUI)
sync_folders:
  - id: 1
    name: "Documents"
    local_path: "/Users/ryan/Documents"
    remote_path: "my-bucket/Documents"
    direction: bidirectional
    enabled: true

    # Sync rules
    excludes:
      - "*.tmp"
      - "*.swp"
      - ".DS_Store"
      - "Thumbs.db"
      - "node_modules/"
      - ".git/"

    # Conflict resolution
    conflict_resolution: keep_local # keep_local, keep_remote, keep_both, manual

    # Scheduling
    sync_mode: continuous # continuous, interval, scheduled
    sync_interval: 300 # seconds (if mode = interval)
    sync_schedule: "*/15 * * * *" # cron format (if mode = scheduled)

    # Limits
    bandwidth_limit: 512 # KB/s per folder
    max_file_size: 5368709120 # 5GB in bytes

  - id: 2
    name: "Photos"
    local_path: "/Users/ryan/Pictures"
    remote_path: "photos-bucket/"
    direction: upload_only
    enabled: true
    excludes:
      - "*.tmp"

# Notification settings
notifications:
  enabled: true
  show_success: false
  show_errors: true
  show_conflicts: true

# API configuration (inherited from main config)
api:
  endpoint: https://api.darkstorage.io
  # api_key loaded from ~/.darkstorage/config.yaml
```

### 5. File Structure

```
darkstorage-cli/
├── cmd/
│   ├── gui/              # GUI application
│   │   ├── main.go       # GUI entry point
│   │   ├── app.go        # Fyne app initialization
│   │   ├── dashboard.go  # Dashboard view
│   │   ├── browser.go    # File browser view
│   │   ├── sync.go       # Sync settings view
│   │   ├── activity.go   # Activity log view
│   │   └── settings.go   # Settings view
│   │
│   ├── daemon/           # Daemon service
│   │   ├── main.go       # Daemon entry point
│   │   ├── daemon.go     # Daemon core logic
│   │   ├── watcher.go    # File system watcher
│   │   ├── syncer.go     # Sync engine
│   │   ├── queue.go      # Operation queue manager
│   │   └── ipc.go        # IPC server
│   │
│   └── [existing CLI commands...]
│
├── internal/
│   ├── api/              # Dark Storage API client
│   │   ├── client.go
│   │   ├── storage.go
│   │   └── auth.go
│   │
│   ├── db/               # Database layer
│   │   ├── db.go         # SQLite connection
│   │   ├── migrations.go # Schema migrations
│   │   ├── folders.go    # Sync folders CRUD
│   │   ├── files.go      # File states CRUD
│   │   ├── queue.go      # Queue operations
│   │   ├── activity.go   # Activity log
│   │   └── conflicts.go  # Conflict management
│   │
│   ├── config/           # Configuration management
│   │   ├── config.go     # Config loading/saving
│   │   └── daemon.go     # Daemon config
│   │
│   ├── sync/             # Sync logic
│   │   ├── engine.go     # Main sync engine
│   │   ├── hasher.go     # File hashing
│   │   ├── differ.go     # State comparison
│   │   ├── conflict.go   # Conflict resolution
│   │   └── excludes.go   # Pattern matching
│   │
│   └── ipc/              # Inter-process communication
│       ├── client.go     # IPC client (for GUI)
│       ├── server.go     # IPC server (for daemon)
│       └── protocol.go   # Message protocol
│
├── go.mod
├── go.sum
├── Makefile              # Updated build targets
└── DESIGN_GUI.md         # This document
```

### 6. Build & Install

**Makefile additions**:
```makefile
# Build GUI application
build-gui:
	go build -o darkstorage-gui cmd/gui/main.go

# Build daemon
build-daemon:
	go build -o darkstorage-daemon cmd/daemon/main.go

# Build all
build-all: build build-gui build-daemon

# Install all binaries
install-all: install
	cp darkstorage-gui /usr/local/bin/
	cp darkstorage-daemon /usr/local/bin/

# Install daemon as system service (systemd)
install-daemon-service:
	cp darkstorage-daemon /usr/local/bin/
	cp scripts/darkstorage-daemon.service /etc/systemd/system/
	systemctl daemon-reload
	systemctl enable darkstorage-daemon

# macOS launchd service
install-daemon-macos:
	cp darkstorage-daemon /usr/local/bin/
	cp scripts/com.darkstorage.daemon.plist ~/Library/LaunchAgents/
	launchctl load ~/Library/LaunchAgents/com.darkstorage.daemon.plist
```

### 7. Dependencies

**New Go modules needed**:
```go
require (
    // Existing...

    // GUI
    fyne.io/fyne/v2 v2.4.0

    // Database
    github.com/mattn/go-sqlite3 v1.14.22

    // File watching
    github.com/fsnotify/fsnotify v1.7.0 // already present

    // Hashing
    // crypto/sha256 from stdlib

    // IPC
    // net package from stdlib for unix sockets

    // Additional utilities
    github.com/robfig/cron/v3 v3.0.1 // for scheduled syncs
)
```

### 8. Security Considerations

1. **API Keys**: Store in system keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service)
2. **File Permissions**: Database and config files should be 0600 (user read/write only)
3. **IPC Socket**: Restrict access to user only
4. **HTTPS**: Always use TLS for API communication
5. **Path Traversal**: Validate all file paths to prevent directory traversal attacks
6. **Input Validation**: Sanitize all user input in GUI

### 9. Future Enhancements

- **Selective sync**: Choose which files/folders within a sync folder
- **Bandwidth scheduling**: Different limits for different times of day
- **Version history**: Browse and restore previous versions via GUI
- **Share management**: Create and manage shares from GUI
- **Mobile notifications**: Push notifications to mobile app
- **Multi-device sync**: Coordinate between multiple devices
- **Encryption**: Client-side encryption option
- **Smart sync**: Machine learning to predict which files to prioritize

## Implementation Phases

### Phase 1: Foundation (Week 1-2)
- Set up project structure
- Implement database layer with migrations
- Create API client wrapper
- Basic configuration management

### Phase 2: Daemon Core (Week 3-4)
- Implement file watcher
- Build sync engine with basic upload/download
- Create operation queue
- Implement IPC server

### Phase 3: GUI Basic (Week 5-6)
- Fyne app skeleton
- Dashboard view with stats
- Settings view for configuration
- IPC client to communicate with daemon

### Phase 4: Full Sync Features (Week 7-8)
- Conflict detection and resolution
- Exclude patterns
- Bidirectional sync
- Error handling and retry logic

### Phase 5: GUI Complete (Week 9-10)
- File browser with drag & drop
- Sync settings management
- Activity log view
- Polish and UX improvements

### Phase 6: Polish & Deploy (Week 11-12)
- System service installation scripts
- Comprehensive testing
- Documentation
- Packaging (DMG for macOS, MSI for Windows, deb/rpm for Linux)

## Testing Strategy

1. **Unit Tests**: All core sync logic, conflict resolution, pattern matching
2. **Integration Tests**: Database operations, API client, file operations
3. **End-to-End Tests**: Complete sync scenarios
4. **Load Tests**: Many files, large files, rapid changes
5. **Platform Tests**: Verify on macOS, Windows, Linux
6. **GUI Tests**: Fyne test framework for UI components
