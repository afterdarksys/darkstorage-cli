package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/darkstorage/cli/internal/db"
	syncpkg "github.com/darkstorage/cli/internal/sync"
	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher       *fsnotify.Watcher
	engine        *syncpkg.Engine
	folders       map[int]*db.SyncFolder
	mu            sync.RWMutex
	debounceDelay time.Duration
	eventBuffer   map[string]*pendingEvent
	bufferMu      sync.Mutex
}

type pendingEvent struct {
	event *syncpkg.FileEvent
	timer *time.Timer
}

func NewWatcher(engine *syncpkg.Engine, debounceDelay time.Duration) (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		watcher:       fw,
		engine:        engine,
		folders:       make(map[int]*db.SyncFolder),
		debounceDelay: debounceDelay,
		eventBuffer:   make(map[string]*pendingEvent),
	}, nil
}

func (w *Watcher) AddFolder(folder *db.SyncFolder) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.watcher.Add(folder.LocalPath); err != nil {
		return err
	}

	w.folders[folder.ID] = folder
	fmt.Printf("Watching: %s\n", folder.LocalPath)
	return nil
}

func (w *Watcher) RemoveFolder(folderID int) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	folder, ok := w.folders[folderID]
	if !ok {
		return nil
	}

	if err := w.watcher.Remove(folder.LocalPath); err != nil {
		return err
	}

	delete(w.folders, folderID)
	return nil
}

func (w *Watcher) Start() {
	go w.eventLoop()
}

func (w *Watcher) Stop() error {
	return w.watcher.Close()
}

func (w *Watcher) eventLoop() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handleEvent(event)

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Watcher error: %v\n", err)
		}
	}
}

func (w *Watcher) handleEvent(event fsnotify.Event) {
	w.mu.RLock()
	var folderID int
	var found bool
	for id, folder := range w.folders {
		if len(event.Name) >= len(folder.LocalPath) && event.Name[:len(folder.LocalPath)] == folder.LocalPath {
			folderID = id
			found = true
			break
		}
	}
	w.mu.RUnlock()

	if !found {
		return
	}

	w.bufferMu.Lock()
	defer w.bufferMu.Unlock()

	if pending, exists := w.eventBuffer[event.Name]; exists {
		pending.timer.Stop()
	}

	eventType := syncpkg.EventModify
	if event.Op&fsnotify.Create != 0 {
		eventType = syncpkg.EventCreate
	} else if event.Op&fsnotify.Remove != 0 {
		eventType = syncpkg.EventDelete
	} else if event.Op&fsnotify.Rename != 0 {
		eventType = syncpkg.EventRename
	}

	fileEvent := &syncpkg.FileEvent{
		Path:      event.Name,
		EventType: eventType,
		Timestamp: time.Now(),
	}

	timer := time.AfterFunc(w.debounceDelay, func() {
		w.processEvent(fileEvent, folderID)
		w.bufferMu.Lock()
		delete(w.eventBuffer, event.Name)
		w.bufferMu.Unlock()
	})

	w.eventBuffer[event.Name] = &pendingEvent{
		event: fileEvent,
		timer: timer,
	}
}

func (w *Watcher) processEvent(event *syncpkg.FileEvent, folderID int) {
	if err := w.engine.ProcessFileEvent(event, folderID); err != nil {
		fmt.Printf("Error processing event: %v\n", err)
	}
}
