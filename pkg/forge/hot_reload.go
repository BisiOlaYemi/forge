package forge

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// HotReloader represents a hot reloader
type HotReloader struct {
	app     *Application
	watcher *fsnotify.Watcher
	cmd     *exec.Cmd
	done    chan bool
}

// NewHotReloader creates a new hot reloader
func NewHotReloader(app *Application) (*HotReloader, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	return &HotReloader{
		app:     app,
		watcher: watcher,
		done:    make(chan bool),
	}, nil
}

// Start starts the hot reloader
func (h *HotReloader) Start() error {
	// Start the application
	if err := h.startApp(); err != nil {
		return err
	}

	// Watch for file changes
	go h.watch()

	return nil
}

// Stop stops the hot reloader
func (h *HotReloader) Stop() error {
	close(h.done)
	if h.cmd != nil && h.cmd.Process != nil {
		if err := h.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}
	return h.watcher.Close()
}

// startApp starts the application
func (h *HotReloader) startApp() error {
	// Kill existing process if any
	if h.cmd != nil && h.cmd.Process != nil {
		if err := h.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill existing process: %w", err)
		}
	}

	// Start new process
	h.cmd = exec.Command("go", "run", ".")
	h.cmd.Stdout = os.Stdout
	h.cmd.Stderr = os.Stderr

	if err := h.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	return nil
}

// watch watches for file changes
func (h *HotReloader) watch() {
	// Watch the current directory and all subdirectories
	if err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		return h.watcher.Add(path)
	}); err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		return
	}

	// Watch for events
	for {
		select {
		case event, ok := <-h.watcher.Events:
			if !ok {
				return
			}

			// Skip non-Go files
			if !strings.HasSuffix(event.Name, ".go") {
				continue
			}

			// Debounce events
			time.Sleep(100 * time.Millisecond)

			// Restart the application
			if err := h.startApp(); err != nil {
				fmt.Printf("Error restarting application: %v\n", err)
			}

		case err, ok := <-h.watcher.Errors:
			if !ok {
				return
			}

			fmt.Printf("Error watching files: %v\n", err)

		case <-h.done:
			return
		}
	}
} 