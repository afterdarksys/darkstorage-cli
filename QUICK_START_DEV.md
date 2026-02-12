# Quick Start - Development Guide

## Get Started in 5 Minutes

This guide will help you start developing the Dark Storage GUI & Daemon immediately.

## Prerequisites Setup

### 1. Install Fyne Dependencies

**macOS:**
```bash
xcode-select --install
```

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install golang gcc libgl1-mesa-dev xorg-dev
```

**Fedora/RHEL:**
```bash
sudo dnf install golang gcc mesa-libGL-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel libXxf86vm-devel
```

**Windows:**
1. Download and install [TDM-GCC](https://jmeubank.github.io/tdm-gcc/)
2. Add to PATH: `C:\TDM-GCC-64\bin`

### 2. Install Go Dependencies

```bash
cd /Users/ryan/development/darkstorage-cli

# Install new dependencies
go get fyne.io/fyne/v2
go get github.com/mattn/go-sqlite3
go get github.com/robfig/cron/v3

# Tidy up
go mod tidy
```

## Project Structure Setup

### Create Directory Structure

```bash
# Create new directories
mkdir -p cmd/gui cmd/daemon
mkdir -p internal/{db,api,sync,ipc,config}
mkdir -p scripts

# Verify structure
tree -L 2 -d
```

Expected output:
```
.
├── cmd
│   ├── daemon
│   └── gui
├── internal
│   ├── api
│   ├── config
│   ├── db
│   ├── ipc
│   └── sync
└── scripts
```

## Phase 1: Database Layer (Day 1)

Start here! This is the foundation.

### Step 1: Create Database Package

```bash
# Create files
touch internal/db/{db.go,migrations.go,models.go,folders.go,files.go}

# Start with db.go
```

**internal/db/db.go** - Minimal working version:
```go
package db

import (
    "database/sql"
    "os"
    "path/filepath"

    _ "github.com/mattn/go-sqlite3"
)

type DB struct {
    conn *sql.DB
    path string
}

func New(dataDir string) (*DB, error) {
    // Ensure directory exists
    if err := os.MkdirAll(dataDir, 0700); err != nil {
        return nil, err
    }

    dbPath := filepath.Join(dataDir, "darkstorage.db")
    conn, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    db := &DB{
        conn: conn,
        path: dbPath,
    }

    // Run migrations
    if err := db.migrate(); err != nil {
        conn.Close()
        return nil, err
    }

    return db, nil
}

func (db *DB) Close() error {
    return db.conn.Close()
}
```

### Step 2: Add Migrations

**internal/db/migrations.go**:
```go
package db

func (db *DB) migrate() error {
    // Create version table
    _, err := db.conn.Exec(`
        CREATE TABLE IF NOT EXISTS schema_version (
            version INTEGER PRIMARY KEY,
            applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `)
    if err != nil {
        return err
    }

    // Check current version
    var version int
    err = db.conn.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
    if err != nil {
        return err
    }

    // Apply migrations
    migrations := []string{
        // Version 1: sync_folders table
        `CREATE TABLE IF NOT EXISTS sync_folders (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            local_path TEXT NOT NULL UNIQUE,
            remote_path TEXT NOT NULL,
            direction TEXT NOT NULL,
            enabled INTEGER DEFAULT 1,
            conflict_resolution TEXT DEFAULT 'keep_local',
            exclude_patterns TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,
        // Add more as you go...
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
```

### Step 3: Test Database

Create a simple test:

```bash
# Create test file
touch internal/db/db_test.go
```

**internal/db/db_test.go**:
```go
package db

import (
    "os"
    "testing"
)

func TestNew(t *testing.T) {
    tmpDir := t.TempDir()

    db, err := New(tmpDir)
    if err != nil {
        t.Fatalf("Failed to create database: %v", err)
    }
    defer db.Close()

    // Verify database exists
    if _, err := os.Stat(db.path); os.IsNotExist(err) {
        t.Error("Database file was not created")
    }
}
```

Run test:
```bash
go test ./internal/db/...
```

## Phase 2: Simple CLI Test (Day 1-2)

Before building the full GUI, create a simple CLI to test the database.

**cmd/daemon/main.go** - Minimal version:
```go
package main

import (
    "fmt"
    "log"
    "os"
    "path/filepath"

    "github.com/darkstorage/cli/internal/db"
)

func main() {
    // Get data directory
    home, err := os.UserHomeDir()
    if err != nil {
        log.Fatal(err)
    }
    dataDir := filepath.Join(home, ".darkstorage")

    // Initialize database
    database, err := db.New(dataDir)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer database.Close()

    fmt.Println("✓ Database initialized successfully")
    fmt.Printf("✓ Location: %s/darkstorage.db\n", dataDir)
}
```

Build and run:
```bash
go run cmd/daemon/main.go
```

Expected output:
```
✓ Database initialized successfully
✓ Location: /Users/ryan/.darkstorage/darkstorage.db
```

## Phase 3: Hello World GUI (Day 2)

Create a minimal Fyne GUI to verify setup.

**cmd/gui/main.go**:
```go
package main

import (
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

func main() {
    myApp := app.New()
    myWindow := myApp.NewWindow("Dark Storage")
    myWindow.Resize(fyne.NewSize(800, 600))

    hello := widget.NewLabel("Dark Storage GUI")
    hello.TextStyle.Bold = true

    content := container.NewVBox(
        hello,
        widget.NewLabel("Status: Not Connected"),
        widget.NewButton("Connect", func() {
            hello.SetText("Connected!")
        }),
    )

    myWindow.SetContent(content)
    myWindow.ShowAndRun()
}
```

Build and run:
```bash
go run cmd/gui/main.go
```

You should see a simple window with a button!

## Update Makefile

Add these targets to your Makefile:

```makefile
# GUI & Daemon builds
.PHONY: build-gui build-daemon build-all

build-gui:
	@echo "Building GUI..."
	go build -o darkstorage-gui cmd/gui/main.go

build-daemon:
	@echo "Building daemon..."
	go build -o darkstorage-daemon cmd/daemon/main.go

build-all: build build-gui build-daemon
	@echo "✓ All binaries built"

# Development helpers
.PHONY: run-gui run-daemon test-all

run-gui: build-gui
	./darkstorage-gui

run-daemon: build-daemon
	./darkstorage-daemon

test-all:
	go test -v ./...

# Clean up
clean:
	rm -f darkstorage darkstorage-gui darkstorage-daemon
```

## Development Workflow

### Daily Development Loop

1. **Morning: Plan**
   ```bash
   # Review checklist
   cat IMPLEMENTATION_CHECKLIST.md

   # Check what you're working on
   # Update your notes
   ```

2. **Development**
   ```bash
   # Work on a feature
   # Test frequently
   make test-all

   # Build and run
   make run-gui
   # or
   make run-daemon
   ```

3. **End of day: Commit**
   ```bash
   git add .
   git commit -m "feat: add database migrations"
   git push
   ```

### Testing Strategy

Test each component as you build it:

```bash
# Test database
go test ./internal/db/... -v

# Test API client
go test ./internal/api/... -v

# Test sync engine
go test ./internal/sync/... -v

# Test everything
make test-all
```

### Debug Tips

**Enable verbose logging:**
```go
// In your code
log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
log.Println("Debug:", value)
```

**SQLite inspection:**
```bash
# Open database
sqlite3 ~/.darkstorage/darkstorage.db

# List tables
.tables

# Show schema
.schema sync_folders

# Query data
SELECT * FROM sync_folders;

# Exit
.quit
```

**Fyne debugging:**
```bash
# Run with debug info
FYNE_THEME=dark go run cmd/gui/main.go

# Run with inspector
go run -tags debug cmd/gui/main.go
```

## Recommended Development Order

### Week 1: Foundation
- [ ] Day 1: Database package (db.go, migrations.go, models.go)
- [ ] Day 2: Database CRUD (folders.go, files.go)
- [ ] Day 3: Configuration package (config.go, daemon.go)
- [ ] Day 4: API client base (client.go, auth.go)
- [ ] Day 5: API storage operations (storage.go)

### Week 2: Sync Core
- [ ] Day 1: File hasher (hasher.go)
- [ ] Day 2: Exclude patterns (excludes.go)
- [ ] Day 3: State differ (differ.go)
- [ ] Day 4: Conflict detection (conflict.go)
- [ ] Day 5: Sync engine core (engine.go)

### Week 3: Daemon
- [ ] Day 1: File watcher with fsnotify
- [ ] Day 2: Queue manager
- [ ] Day 3: IPC protocol
- [ ] Day 4: IPC server
- [ ] Day 5: Daemon integration

### Week 4: GUI
- [ ] Day 1: App structure & navigation
- [ ] Day 2: Dashboard view
- [ ] Day 3: File browser view
- [ ] Day 4: Sync settings view
- [ ] Day 5: Activity log & polish

## Common Issues & Solutions

### Issue: "package fyne.io/fyne/v2: no Go files"

**Solution:**
```bash
go clean -modcache
go get fyne.io/fyne/v2
go mod tidy
```

### Issue: "undefined: sqlite3"

**Solution:**
```bash
# Make sure you have gcc
gcc --version

# Reinstall sqlite3
go get -u github.com/mattn/go-sqlite3
```

### Issue: GUI window doesn't show (Linux)

**Solution:**
```bash
# Install X11 dependencies
sudo apt-get install libgl1-mesa-dev xorg-dev
```

## Getting Help

1. **Fyne Documentation**: https://developer.fyne.io/
2. **SQLite3 Go Driver**: https://github.com/mattn/go-sqlite3
3. **fsnotify**: https://github.com/fsnotify/fsnotify

## Next Steps

Once you've completed the quick start:

1. Review `DESIGN_GUI.md` for full architecture
2. Follow `IMPLEMENTATION_CHECKLIST.md` step by step
3. Use `DATA_STRUCTURES.md` as reference for types

Start building! The best way to learn is to code.

## Example Session

Here's what a typical development session looks like:

```bash
# 1. Start your day
cd /Users/ryan/development/darkstorage-cli
git pull

# 2. Check what you're working on
cat IMPLEMENTATION_CHECKLIST.md | grep "Day 1"

# 3. Create your files
touch internal/db/folders.go

# 4. Write code...
# (use your editor)

# 5. Test frequently
go test ./internal/db/... -v

# 6. Run it
make run-daemon

# 7. Commit when working
git add internal/db/folders.go
git commit -m "feat(db): add sync folder CRUD operations"
git push

# 8. Take a break!
```

Good luck! 🚀
