package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/darkstorage/cli/internal/api"
	"github.com/darkstorage/cli/internal/config"
	"github.com/darkstorage/cli/internal/db"
	"github.com/darkstorage/cli/internal/ipc"
	syncpkg "github.com/darkstorage/cli/internal/sync"
	"github.com/spf13/viper"
)

type Daemon struct {
	db        *db.DB
	client    *api.Client
	engine    *syncpkg.Engine
	watcher   *Watcher
	ipcServer *ipc.Server
	config    *config.DaemonConfig
	startTime time.Time
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "start":
			runDaemon()
		case "stop":
			stopDaemon()
		case "status":
			daemonStatus()
		default:
			fmt.Println("Usage: darkstorage-daemon {start|stop|status}")
			os.Exit(1)
		}
	} else {
		runDaemon()
	}
}

func runDaemon() {
	dataDir, err := config.GetDefaultDataDir()
	if err != nil {
		log.Fatalf("Failed to get data directory: %v", err)
	}

	cfg, err := config.LoadDaemonConfig()
	if err != nil {
		log.Printf("Failed to load config, using defaults: %v", err)
		cfg = &config.DaemonConfig{}
	}

	database, err := db.New(dataDir)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	apiKey := viper.GetString("api_key")
	endpoint := viper.GetString("endpoint")
	if endpoint == "" {
		endpoint = "https://api.darkstorage.io"
	}

	client := api.NewClient(endpoint, apiKey)
	engine := syncpkg.NewEngine(database, client)

	socketPath := filepath.Join(dataDir, "daemon.sock")
	ipcServer := ipc.NewServer(socketPath)

	daemon := &Daemon{
		db:        database,
		client:    client,
		engine:    engine,
		ipcServer: ipcServer,
		config:    cfg,
		startTime: time.Now(),
	}

	daemon.setupIPCHandlers()

	if err := ipcServer.Start(); err != nil {
		log.Fatalf("Failed to start IPC server: %v", err)
	}
	defer ipcServer.Stop()

	watcher, err := NewWatcher(engine, syncpkg.DefaultDebounceDelay)
	if err != nil {
		log.Fatalf("Failed to create watcher: %v", err)
	}
	daemon.watcher = watcher
	defer watcher.Stop()

	folders, err := database.ListSyncFolders()
	if err != nil {
		log.Fatalf("Failed to list sync folders: %v", err)
	}

	for _, folder := range folders {
		if folder.Enabled {
			if err := watcher.AddFolder(folder); err != nil {
				log.Printf("Failed to watch folder %s: %v", folder.LocalPath, err)
			}
		}
	}

	watcher.Start()

	go daemon.queueWorker()

	fmt.Printf("Dark Storage daemon started\n")
	fmt.Printf("IPC socket: %s\n", socketPath)
	fmt.Printf("Watching %d folder(s)\n", len(folders))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
}

func (d *Daemon) setupIPCHandlers() {
	d.ipcServer.RegisterHandler("status", d.handleStatus)
	d.ipcServer.RegisterHandler("add_sync_folder", d.handleAddSyncFolder)
	d.ipcServer.RegisterHandler("remove_sync_folder", d.handleRemoveSyncFolder)
	d.ipcServer.RegisterHandler("get_activity", d.handleGetActivity)
	d.ipcServer.RegisterHandler("force_sync", d.handleForceSync)
	d.ipcServer.RegisterHandler("get_config", d.handleGetConfig)
	d.ipcServer.RegisterHandler("set_config", d.handleSetConfig)
}

func (d *Daemon) handleStatus(data json.RawMessage) (*ipc.Response, error) {
	folders, err := d.db.ListSyncFolders()
	if err != nil {
		return nil, err
	}

	queueSize, err := d.db.GetQueueSize()
	if err != nil {
		return nil, err
	}

	var folderStatuses []ipc.SyncFolderStatus
	for _, folder := range folders {
		folderStatuses = append(folderStatuses, ipc.SyncFolderStatus{
			ID:         folder.ID,
			LocalPath:  folder.LocalPath,
			RemotePath: folder.RemotePath,
			Status:     "idle",
		})
	}

	status := &ipc.StatusResponse{
		DaemonRunning: true,
		SyncFolders:   folderStatuses,
		QueueSize:     queueSize,
		Uptime:        time.Since(d.startTime).String(),
	}

	responseData, err := json.Marshal(status)
	if err != nil {
		return nil, err
	}

	return &ipc.Response{
		Success: true,
		Data:    responseData,
	}, nil
}

func (d *Daemon) handleAddSyncFolder(data json.RawMessage) (*ipc.Response, error) {
	var req ipc.AddSyncFolderRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}

	folder := &db.SyncFolder{
		LocalPath:          req.LocalPath,
		RemotePath:         req.RemotePath,
		Direction:          req.Direction,
		Enabled:            true,
		ConflictResolution: req.ConflictResolution,
	}

	if err := d.db.CreateSyncFolder(folder); err != nil {
		return nil, err
	}

	if err := d.watcher.AddFolder(folder); err != nil {
		return nil, err
	}

	result := &ipc.AddSyncFolderResponse{ID: folder.ID}
	responseData, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return &ipc.Response{
		Success: true,
		Data:    responseData,
	}, nil
}

func (d *Daemon) handleRemoveSyncFolder(data json.RawMessage) (*ipc.Response, error) {
	var req ipc.RemoveSyncFolderRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}

	if err := d.watcher.RemoveFolder(req.ID); err != nil {
		return nil, err
	}

	if err := d.db.DeleteSyncFolder(req.ID); err != nil {
		return nil, err
	}

	return &ipc.Response{Success: true}, nil
}

func (d *Daemon) handleGetActivity(data json.RawMessage) (*ipc.Response, error) {
	var req ipc.GetActivityRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}

	if req.Limit == 0 {
		req.Limit = 50
	}

	activities, err := d.db.GetRecentActivity(req.Limit)
	if err != nil {
		return nil, err
	}

	var entries []ipc.ActivityEntry
	for _, activity := range activities {
		entry := ipc.ActivityEntry{
			ID:        activity.ID,
			Operation: activity.Operation,
			Path:      activity.Path,
			Status:    activity.Status,
			Timestamp: activity.CreatedAt,
		}
		if activity.BytesTransferred != nil {
			entry.BytesTransferred = *activity.BytesTransferred
		}
		if activity.ErrorMessage != nil {
			entry.ErrorMessage = *activity.ErrorMessage
		}
		entries = append(entries, entry)
	}

	result := &ipc.GetActivityResponse{Activities: entries}
	responseData, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return &ipc.Response{
		Success: true,
		Data:    responseData,
	}, nil
}

func (d *Daemon) handleForceSync(data json.RawMessage) (*ipc.Response, error) {
	var req ipc.ForceSyncRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}

	go d.engine.SyncFolder(req.FolderID)

	return &ipc.Response{Success: true}, nil
}

func (d *Daemon) handleGetConfig(data json.RawMessage) (*ipc.Response, error) {
	configMap := map[string]interface{}{
		"daemon":        d.config.Daemon,
		"notifications": d.config.Notifications,
		"api":           d.config.API,
	}

	result := &ipc.GetConfigResponse{Config: configMap}
	responseData, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return &ipc.Response{
		Success: true,
		Data:    responseData,
	}, nil
}

func (d *Daemon) handleSetConfig(data json.RawMessage) (*ipc.Response, error) {
	var req ipc.SetConfigRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}

	if err := config.SaveDaemonConfig(d.config); err != nil {
		return nil, err
	}

	return &ipc.Response{Success: true}, nil
}

func (d *Daemon) queueWorker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := d.engine.ProcessQueue(); err != nil {
			log.Printf("Queue processing error: %v", err)
		}
	}
}

func stopDaemon() {
	fmt.Println("Stopping daemon...")
}

func daemonStatus() {
	dataDir, err := config.GetDefaultDataDir()
	if err != nil {
		log.Fatalf("Failed to get data directory: %v", err)
	}

	socketPath := filepath.Join(dataDir, "daemon.sock")
	client := ipc.NewClient(socketPath)

	status, err := client.GetStatus()
	if err != nil {
		fmt.Printf("Daemon is not running\n")
		return
	}

	fmt.Printf("Daemon Status: Running\n")
	fmt.Printf("Uptime: %s\n", status.Uptime)
	fmt.Printf("Queue Size: %d\n", status.QueueSize)
	fmt.Printf("Sync Folders: %d\n", len(status.SyncFolders))
}
