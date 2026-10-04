package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"reflect"
	"runtime"
	"syscall"
	"time"
)

type fileState struct {
	modified time.Time
	size     int64
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := watch(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}

func watch(ctx context.Context) error {
	binaryPath := filepath.Join(os.TempDir(), "todoapp-dev")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}
	defer os.Remove(binaryPath)

	current, err := snapshot(".")
	if err != nil {
		return fmt.Errorf("scan source files: %w", err)
	}
	var server *exec.Cmd
	var serverDone chan error
	ticker := time.NewTicker(600 * time.Millisecond)
	defer ticker.Stop()

	start := func() {
		build := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, "./cmd/server")
		build.Stdout = os.Stdout
		build.Stderr = os.Stderr
		if err := build.Run(); err != nil {
			log.Printf("build failed: %v", err)
			return
		}
		server = exec.Command(binaryPath)
		server.Stdout = os.Stdout
		server.Stderr = os.Stderr
		if err := server.Start(); err != nil {
			log.Printf("start server: %v", err)
			server = nil
			return
		}
		serverDone = make(chan error, 1)
		go func() { serverDone <- server.Wait() }()
	}
	start()

	for {
		select {
		case <-ctx.Done():
			if server != nil {
				_ = server.Process.Kill()
				<-serverDone
			}
			return nil
		case err := <-serverDone:
			if server != nil {
				log.Printf("server stopped: %v", err)
			}
			server, serverDone = nil, nil
		case <-ticker.C:
			next, err := snapshot(".")
			if err != nil {
				return fmt.Errorf("scan source files: %w", err)
			}
			if reflect.DeepEqual(current, next) {
				continue
			}
			current = next
			if server != nil {
				log.Print("source change detected; restarting server")
				_ = server.Process.Kill()
				<-serverDone
				server, serverDone = nil, nil
			} else {
				log.Print("source change detected; rebuilding server")
			}
			if ctx.Err() == nil {
				start()
			}
		}
	}
}

func snapshot(root string) (map[string]fileState, error) {
	files := make(map[string]fileState)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != root && (entry.Name() == ".git" || entry.Name() == "bin" || entry.Name() == "golang-todoapp" || entry.Name() == "node_modules") {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(entry.Name())
		if ext != ".go" && ext != ".html" && ext != ".js" && ext != ".css" && ext != ".mod" && ext != ".sum" && ext != ".yml" && ext != ".yaml" && entry.Name() != "Makefile" && entry.Name() != ".env" {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		files[path] = fileState{modified: info.ModTime(), size: info.Size()}
		return nil
	})
	return files, err
}
