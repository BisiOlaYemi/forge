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


type HotReloader struct {
	app     *Application
	watcher *fsnotify.Watcher
	cmd     *exec.Cmd
	done    chan bool
}


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


func (h *HotReloader) Start() error {
	if err := h.startApp(); err != nil {
		return err
	}

	
	go h.watch()

	return nil
}

func (h *HotReloader) Stop() error {
	close(h.done)
	if h.cmd != nil && h.cmd.Process != nil {
		if err := h.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}
	return h.watcher.Close()
}

func (h *HotReloader) startApp() error {
	if h.cmd != nil && h.cmd.Process != nil {
		if err := h.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill existing process: %w", err)
		}
	}

	h.cmd = exec.Command("go", "run", ".")
	h.cmd.Stdout = os.Stdout
	h.cmd.Stderr = os.Stderr

	if err := h.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	return nil
}

func (h *HotReloader) watch() {
	if err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		return h.watcher.Add(path)
	}); err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		return
	}

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

			// Debounce 
			time.Sleep(100 * time.Millisecond)

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