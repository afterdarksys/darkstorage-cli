# Dark Storage GUI & Daemon - Implementation Checklist

## Prerequisites

- [ ] Install Fyne prerequisites for your OS:
  - macOS: `xcode-select --install`
  - Linux: `sudo apt-get install golang gcc libgl1-mesa-dev xorg-dev`
  - Windows: Install TDM-GCC

- [ ] Add new dependencies to go.mod:
```bash
go get fyne.io/fyne/v2
go get github.com/mattn/go-sqlite3
go get github.com/robfig/cron/v3
```

## Step-by-Step Implementation

### 1. Database Layer (Start Here)

- [ ] `internal/db/db.go` - Create SQLite connection manager
  - Initialize database connection
  - Run migrations on first startup
  - Connection pooling

- [ ] `internal/db/migrations.go` - Schema migrations
  - Implement migration runner
  - Add all tables from DESIGN_GUI.md schema
  - Version tracking for migrations

- [ ] `internal/db/folders.go` - Sync folders CRUD
  - CreateSyncFolder(folder SyncFolder) error
  - GetSyncFolder(id int) (*SyncFolder, error)
  - ListSyncFolders() ([]SyncFolder, error)
  - UpdateSyncFolder(folder SyncFolder) error
  - DeleteSyncFolder(id int) error

- [ ] `internal/db/files.go` - File states CRUD
  - UpsertFileState(state FileState) error
  - GetFileState(folderID int, path string) (*FileState, error)
  - ListFileStates(folderID int) ([]FileState, error)
  - UpdateSyncStatus(id int, status string) error

- [ ] `internal/db/queue.go` - Queue operations
  - EnqueueOperation(op QueueOperation) error
  - DequeueOperation() (*QueueOperation, error)
  - UpdateOperationStatus(id int, status string) error
  - GetQueueSize() int

- [ ] `internal/db/activity.go` - Activity logging
  - LogActivity(activity Activity) error
  - GetRecentActivity(limit int) ([]Activity, error)
  - GetActivityByFolder(folderID int, limit int) ([]Activity, error)

- [ ] `internal/db/conflicts.go` - Conflict management
  - CreateConflict(conflict Conflict) error
  - GetUnresolvedConflicts() ([]Conflict, error)
  - ResolveConflict(id int, resolution string) error

### 2. Configuration Management

- [ ] `internal/config/daemon.go` - Daemon configuration
  - Define DaemonConfig struct matching daemon.yaml
  - LoadDaemonConfig() (*DaemonConfig, error)
  - SaveDaemonConfig(config *DaemonConfig) error
  - ValidateConfig() error

- [ ] Extend existing `cmd/config.go` for daemon settings
  - Add daemon-specific config commands
  - Merge with existing config system

### 3. API Client Wrapper

- [ ] `internal/api/client.go` - Base client
  - NewClient(endpoint, apiKey string) *Client
  - SetTimeout(duration time.Duration)
  - HandleRetries with exponential backoff

- [ ] `internal/api/storage.go` - Storage operations
  - UploadFile(localPath, remotePath string) error
  - DownloadFile(remotePath, localPath string) error
  - DeleteFile(remotePath string) error
  - ListFiles(remotePath string) ([]FileInfo, error)
  - GetFileMetadata(remotePath string) (*FileMetadata, error)

- [ ] `internal/api/auth.go` - Authentication
  - Login(username, password string) (string, error)
  - ValidateToken(token string) error
  - RefreshToken() error

### 4. Sync Engine Core

- [ ] `internal/sync/hasher.go` - File hashing
  - HashFile(path string) (string, error) // SHA-256
  - HashFileChunked(path string, chunkSize int) (string, error)
  - CompareHashes(hash1, hash2 string) bool

- [ ] `internal/sync/excludes.go` - Pattern matching
  - CompilePatterns(patterns []string) (*ExcludeRules, error)
  - ShouldExclude(path string, rules *ExcludeRules) bool
  - Support glob patterns (*.tmp, node_modules/, etc.)

- [ ] `internal/sync/differ.go` - State comparison
  - CompareStates(local, remote FileState) SyncAction
  - DetermineConflict(local, remote FileState) bool
  - CalculateSyncActions(states []FileState) []SyncAction

- [ ] `internal/sync/conflict.go` - Conflict resolution
  - ResolveConflict(conflict Conflict, strategy string) (Action, error)
  - Strategies: keep_local, keep_remote, keep_both, manual

- [ ] `internal/sync/engine.go` - Main sync engine
  - Initialize(db *sql.DB, api *api.Client, config *Config)
  - SyncFolder(folderID int) error
  - ProcessFileEvent(event FileEvent) error
  - HandleUpload(localPath, remotePath string) error
  - HandleDownload(remotePath, localPath string) error
  - HandleDelete(path string, isLocal bool) error

### 5. File System Watcher

- [ ] `cmd/daemon/watcher.go` - fsnotify wrapper
  - NewWatcher(folders []SyncFolder) (*Watcher, error)
  - AddFolder(path string) error
  - RemoveFolder(path string) error
  - Start() error
  - Stop() error
  - Events channel: chan FileEvent
  - Debounce logic (wait 3s after last change)

### 6. Operation Queue Manager

- [ ] `cmd/daemon/queue.go` - Queue management
  - NewQueueManager(db *sql.DB, workers int)
  - Start() error
  - Stop() error
  - Enqueue(operation Operation) error
  - Worker pool with configurable concurrency
  - Retry logic with exponential backoff
  - Priority handling (delete > upload > download)

### 7. IPC Layer

- [ ] `internal/ipc/protocol.go` - Message protocol
  - Define all command/response structs
  - Marshal/unmarshal JSON messages
  - Command types: status, add_folder, remove_folder, pause, resume, etc.

- [ ] `internal/ipc/server.go` - IPC server (daemon side)
  - NewServer(socketPath string) (*Server, error)
  - Start() error
  - Stop() error
  - HandleConnection(conn net.Conn)
  - RegisterHandler(command string, handler func)

- [ ] `internal/ipc/client.go` - IPC client (GUI side)
  - NewClient(socketPath string) (*Client, error)
  - Connect() error
  - SendCommand(command Command) (*Response, error)
  - Close() error

### 8. Daemon Service

- [ ] `cmd/daemon/daemon.go` - Main daemon logic
  - Initialize all components (db, api, sync engine, watcher, queue, ipc)
  - Start() error
  - Stop() error
  - Reload() error (reload config without restart)
  - Signal handling (SIGTERM, SIGHUP, SIGINT)
  - PID file management

- [ ] `cmd/daemon/syncer.go` - Sync coordinator
  - Coordinates between watcher events and sync engine
  - Batch processing of events
  - Rate limiting

- [ ] `cmd/daemon/main.go` - Entry point
  - Parse command line flags (start, stop, restart, status)
  - Daemonize process (fork on Unix)
  - Logging setup

### 9. Fyne GUI Application

- [ ] `cmd/gui/app.go` - App initialization
  - Initialize Fyne app
  - Create main window
  - Setup IPC client connection to daemon
  - Load configuration
  - Setup system tray icon

- [ ] `cmd/gui/dashboard.go` - Dashboard view
  - Storage usage chart (Fyne canvas)
  - Recent activity list
  - Quick stats widgets
  - Sync status indicators
  - Auto-refresh every 5 seconds

- [ ] `cmd/gui/browser.go` - File browser view
  - Tree widget for bucket/folder structure
  - File list table
  - Context menu (upload, download, delete, share)
  - Drag & drop for uploads
  - Search/filter bar
  - Multi-select support

- [ ] `cmd/gui/sync.go` - Sync settings view
  - List of sync folders (Table or List)
  - Add folder button → dialog with folder picker
  - Remove folder button with confirmation
  - Edit folder settings:
    - Sync direction dropdown
    - Exclude patterns text area
    - Conflict resolution dropdown
    - Bandwidth limit entry
    - Sync schedule radio buttons
  - Test sync button (run manual sync)

- [ ] `cmd/gui/activity.go` - Activity log view
  - Scrollable list of activity entries
  - Filter by type dropdown
  - Date range picker
  - Export button (save as CSV/JSON)
  - Auto-refresh with new entries highlighted

- [ ] `cmd/gui/settings.go` - Settings view
  - API configuration (endpoint, key)
  - Daemon control buttons (start, stop, restart, status)
  - Notification preferences checkboxes
  - Log level dropdown
  - Cache location picker
  - About section (version, license, credits)

- [ ] `cmd/gui/main.go` - GUI entry point
  - Create Fyne app
  - Setup theme
  - Create main window with tabs/navigation
  - Launch

### 10. System Integration

- [ ] `scripts/darkstorage-daemon.service` - systemd service file
```ini
[Unit]
Description=Dark Storage Sync Daemon
After=network.target

[Service]
Type=simple
User=%i
ExecStart=/usr/local/bin/darkstorage-daemon start
ExecStop=/usr/local/bin/darkstorage-daemon stop
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

- [ ] `scripts/com.darkstorage.daemon.plist` - macOS LaunchAgent
```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.darkstorage.daemon</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/darkstorage-daemon</string>
        <string>start</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
```

- [ ] Update Makefile with new build targets (from DESIGN_GUI.md)

### 11. Testing

- [ ] Unit tests for sync engine logic
  - Test conflict detection
  - Test exclude patterns
  - Test hash calculation
  - Test sync action determination

- [ ] Integration tests
  - Test database operations
  - Test API client with mock server
  - Test file operations

- [ ] End-to-end tests
  - Create test sync scenario
  - Modify files
  - Verify sync completion
  - Test conflict scenarios

- [ ] GUI tests
  - Fyne test framework
  - Test navigation
  - Test form validation

### 12. Documentation

- [ ] Update README.md with GUI installation instructions
- [ ] Create USER_GUIDE.md for GUI usage
- [ ] Add daemon configuration examples
- [ ] Document API for IPC protocol
- [ ] Add troubleshooting section

### 13. Packaging & Distribution

- [ ] macOS: Create .app bundle and DMG
- [ ] Windows: Create MSI installer
- [ ] Linux: Create .deb and .rpm packages
- [ ] Docker: Optional containerized daemon
- [ ] Release automation (GitHub Actions / GitLab CI)

## Development Order Recommendation

Start with this order for fastest progress:

1. **Database layer first** - Foundation for everything
2. **Config management** - Need to load settings
3. **API client** - Can test with existing CLI code
4. **Sync engine core** (hasher, differ, excludes) - Business logic
5. **Simple daemon** without watcher - Test sync manually
6. **File watcher** - Add automatic detection
7. **Queue manager** - Add proper job handling
8. **IPC** - Enable GUI communication
9. **Basic GUI** with dashboard and settings - User can interact
10. **Complete GUI** - All views fully functional
11. **Polish & testing** - Make it production ready

## Quick Start Commands

After implementing the basics:

```bash
# Build everything
make build-all

# Initialize database (first run)
./darkstorage-daemon init

# Start daemon
./darkstorage-daemon start

# Launch GUI
./darkstorage-gui

# Check daemon status
./darkstorage-daemon status

# Stop daemon
./darkstorage-daemon stop
```

## Key Design Decisions to Make

Before implementation, decide:

- [ ] **Sync conflict strategy default**: keep_local, keep_remote, or manual?
- [ ] **Queue persistence**: In-memory or database? (Recommend database for reliability)
- [ ] **Change detection**: Hash-based or timestamp-based? (Hash more reliable)
- [ ] **API authentication**: Token refresh strategy?
- [ ] **Error notification**: Desktop notifications or in-app only?
- [ ] **Log rotation**: Max size and age for log files?
- [ ] **Multi-instance**: Allow multiple daemons or single instance per user?

## Performance Targets

- [ ] Support at least 10,000 files per sync folder
- [ ] Upload/download speeds limited only by network (when no limit set)
- [ ] GUI should remain responsive (60 FPS) even during heavy sync
- [ ] Database queries < 100ms for typical operations
- [ ] File change detection latency < 5 seconds
- [ ] Memory usage < 200 MB for daemon with 5 sync folders

## Security Checklist

- [ ] Store API keys in system keychain (not plaintext)
- [ ] Validate all file paths (prevent traversal attacks)
- [ ] Set proper permissions on database (0600)
- [ ] Set proper permissions on IPC socket (user only)
- [ ] Use HTTPS for all API calls
- [ ] Validate all user input in GUI forms
- [ ] No credentials in logs
- [ ] Secure cleanup on uninstall (offer to keep or delete data)
