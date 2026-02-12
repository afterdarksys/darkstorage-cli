# Dark Storage GUI & Daemon

Full-featured GUI application and sync daemon for Dark Storage.

## Quick Start

### 1. Build Everything

```bash
make build-all-bins
```

This creates three binaries:
- `darkstorage` - CLI tool
- `darkstorage-gui` - GUI application
- `darkstorage-daemon` - Background sync daemon

### 2. Start the Daemon

```bash
./darkstorage-daemon start
```

Or run in foreground:
```bash
./darkstorage-daemon
```

### 3. Launch the GUI

```bash
./darkstorage-gui
```

## Features

### Daemon
- Automatic file synchronization
- File system watching with fsnotify
- SQLite database for state tracking
- Queue-based operation processing with retry logic
- IPC server for GUI communication
- Configurable sync rules per folder

### GUI (Fyne)
- Dashboard view with status and activity
- Sync folder management
- Activity log viewer
- Settings configuration
- Real-time updates every 5 seconds
- Add/remove sync folders via UI
- View and edit daemon configuration

## Architecture

```
GUI ←→ IPC Socket ←→ Daemon ←→ Database
                       ↓
                   File Watcher
                       ↓
                   Sync Engine
                       ↓
                   API Client
```

## Configuration

Config file: `~/.darkstorage/daemon.yaml`

Example:
```yaml
daemon:
  enabled: true
  log_level: info
  worker_threads: 4
  debounce_delay: 3s

notifications:
  enabled: true
  show_errors: true
  show_conflicts: true

api:
  endpoint: https://api.darkstorage.io
```

## Database

Location: `~/.darkstorage/darkstorage.db`

Tables:
- `sync_folders` - Watched directories configuration
- `file_states` - Per-file sync state
- `sync_queue` - Pending operations
- `activity_log` - Operation history
- `conflicts` - Detected conflicts

## Usage

### Add a Sync Folder (GUI)
1. Launch GUI
2. Click "Sync Folders"
3. Click "Add Folder"
4. Enter local path, remote path, and sync direction
5. Click Submit

### Check Daemon Status

```bash
./darkstorage-daemon status
```

### View Logs

```bash
tail -f ~/.darkstorage/daemon.log
```

### Stop Daemon

```bash
./darkstorage-daemon stop
```

## Development

### Run Tests

```bash
make test
```

### Build Individual Components

```bash
make build-gui      # GUI only
make build-daemon   # Daemon only
make build          # CLI only
```

### Run Without Building

```bash
make run-gui
make run-daemon
```

## Installation

Install all binaries system-wide:

```bash
sudo make install-all
```

This installs to `/usr/local/bin/`:
- `darkstorage`
- `darkstorage-gui`
- `darkstorage-daemon`

## Troubleshooting

### GUI won't start
Make sure Fyne dependencies are installed (see QUICK_START_DEV.md)

### Daemon won't connect
Check that the daemon is running:
```bash
./darkstorage-daemon status
```

### Database errors
Remove and reinitialize:
```bash
rm ~/.darkstorage/darkstorage.db
./darkstorage-daemon
```

### IPC socket issues
Remove stale socket:
```bash
rm ~/.darkstorage/daemon.sock
./darkstorage-daemon start
```

## Features Implemented

✅ SQLite database with full schema
✅ Database CRUD operations
✅ Configuration management (YAML)
✅ API client with progress tracking
✅ File hasher (SHA-256)
✅ Exclude pattern matching
✅ Sync engine with queue processing
✅ File system watcher with debouncing
✅ IPC protocol (Unix socket)
✅ IPC server and client
✅ Daemon with all handlers
✅ Fyne GUI with 4 main views
✅ Real-time status updates
✅ Add/remove sync folders from GUI
✅ Activity log viewer
✅ Configuration viewer/editor
✅ Makefile with all targets

## Next Steps

See DESIGN_GUI.md and IMPLEMENTATION_CHECKLIST.md for:
- Conflict resolution UI
- Bandwidth limiting
- Scheduled syncing
- Client-side encryption
- System service installation
- Cross-platform packaging
