package transfer

import (
	"context"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type fileState struct {
	size        int64
	modTime     time.Time
	stableSince time.Time
}

// Watcher monitors the Send directory for stable files.
type Watcher struct {
	sendDir     string
	watcher     *fsnotify.Watcher
	onFileReady func(absPath, relPath, peerName string)

	mu           sync.Mutex
	pendingFiles map[string]*fileState
}

// NewWatcher creates a new Watcher for the Send directory.
func NewWatcher(sendDir string, onFileReady func(absPath, relPath, peerName string)) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		sendDir:      sendDir,
		watcher:      w,
		onFileReady:  onFileReady,
		pendingFiles: make(map[string]*fileState),
	}, nil
}

// AddDirectory manually adds a directory to be watched by fsnotify.
func (w *Watcher) AddDirectory(path string) error {
	return w.watcher.Add(path)
}

// Start begins watching the directory tree.
func (w *Watcher) Start(ctx context.Context) error {
	// Ensure Send directory exists
	if err := os.MkdirAll(w.sendDir, 0755); err != nil {
		return err
	}

	// Add watches for all existing directories
	err := filepath.WalkDir(w.sendDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return w.watcher.Add(path)
		}
		// Also track existing files that might not have been sent
		w.trackFile(path)
		return nil
	})
	if err != nil {
		return err
	}

	go w.eventLoop(ctx)
	go w.stabilityLoop(ctx)

	log.Printf("Started watching directory: %s", w.sendDir)
	return nil
}

func (w *Watcher) eventLoop(ctx context.Context) {
	defer w.watcher.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
				info, err := os.Stat(event.Name)
				if err != nil {
					continue
				}
				if info.IsDir() {
					w.watcher.Add(event.Name) // Watch new directory
				} else {
					w.trackFile(event.Name) // Track file for stability
				}
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

func (w *Watcher) trackFile(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return
	}

	// If it's a temp file, skip it
	if strings.HasSuffix(info.Name(), ".part") {
		return
	}

	state, exists := w.pendingFiles[path]
	if !exists {
		w.pendingFiles[path] = &fileState{
			size:        info.Size(),
			modTime:     info.ModTime(),
			stableSince: time.Now(),
		}
	} else {
		// Update state
		if state.size != info.Size() || state.modTime != info.ModTime() {
			state.size = info.Size()
			state.modTime = info.ModTime()
			state.stableSince = time.Now()
		}
	}
}

func (w *Watcher) stabilityLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	stabilityDuration := 2 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.checkStability(stabilityDuration)
		}
	}
}

func (w *Watcher) checkStability(duration time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	for path, state := range w.pendingFiles {
		// Re-stat to ensure it hasn't changed since last tracked
		info, err := os.Stat(path)
		if err != nil {
			// File disappeared or can't be read
			delete(w.pendingFiles, path)
			continue
		}

		if info.Size() != state.size || info.ModTime() != state.modTime {
			// File is still changing
			state.size = info.Size()
			state.modTime = info.ModTime()
			state.stableSince = now
			continue
		}

		if now.Sub(state.stableSince) >= duration {
			// File is stable!
			delete(w.pendingFiles, path)
			w.dispatchFile(path)
		}
	}
}

func (w *Watcher) dispatchFile(path string) {
	// Parse peer name and relative path
	// SendDir is like /.../LanShare/Send
	// Path is like /.../LanShare/Send/PeerB/photos/img.png
	relToRoot, err := filepath.Rel(w.sendDir, path)
	if err != nil {
		return
	}

	parts := strings.Split(filepath.ToSlash(relToRoot), "/")
	if len(parts) < 2 {
		return // File is directly in Send/ instead of Send/PeerName/
	}

	peerName := parts[0]
	relPath := strings.Join(parts[1:], "/")

	if w.onFileReady != nil {
		go w.onFileReady(path, relPath, peerName)
	}
}
