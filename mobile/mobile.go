package mobile

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jzxq45/golang-todoapp/internal/httpapp"
	"github.com/jzxq45/golang-todoapp/internal/task"
)

var version = "dev"

type runningServer struct {
	dataDir  string
	port     int
	server   *http.Server
	logger   *slog.Logger
	logFile  *os.File
	serveErr chan error
}

var serverState struct {
	sync.Mutex
	running *runningServer
}

func Start(dataDir string) (int, error) {
	if dataDir == "" {
		return 0, errors.New("application data directory is empty")
	}
	absoluteDir, err := filepath.Abs(dataDir)
	if err != nil {
		return 0, fmt.Errorf("resolve application data directory: %w", err)
	}

	serverState.Lock()
	defer serverState.Unlock()
	if serverState.running != nil {
		select {
		case err := <-serverState.running.serveErr:
			running := serverState.running
			serverState.running = nil
			if closeErr := running.logFile.Close(); closeErr != nil {
				return 0, fmt.Errorf("close stopped mobile application log: %w", closeErr)
			}
			if err != nil {
				return 0, fmt.Errorf("mobile HTTP server stopped unexpectedly: %w", err)
			}
			return 0, errors.New("mobile HTTP server is stopped")
		default:
		}
		if serverState.running.dataDir != absoluteDir {
			return 0, errors.New("mobile server is already running with a different data directory")
		}
		return serverState.running.port, nil
	}

	if err := os.MkdirAll(absoluteDir, 0o700); err != nil {
		return 0, fmt.Errorf("create application data directory: %w", err)
	}
	logFile, err := os.OpenFile(filepath.Join(absoluteDir, "app.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return 0, fmt.Errorf("open application log: %w", err)
	}
	logger := slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: slog.LevelInfo}))

	repository, err := task.NewJSONRepository(filepath.Join(absoluteDir, "tasks.json"))
	if err != nil {
		return 0, errors.Join(
			fmt.Errorf("open task repository: %w", err),
			wrapIfError("close mobile application log", logFile.Close()),
		)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, errors.Join(
			fmt.Errorf("listen on local HTTP port: %w", err),
			wrapIfError("close mobile application log", logFile.Close()),
		)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	server := &http.Server{
		Addr:              listener.Addr().String(),
		Handler:           httpapp.New(task.NewHandler(task.NewService(repository)), logger),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	running := &runningServer{
		dataDir:  absoluteDir,
		port:     port,
		server:   server,
		logger:   logger,
		logFile:  logFile,
		serveErr: make(chan error, 1),
	}
	serverState.running = running
	logger.Info("mobile HTTP server started", "version", version, "address", server.Addr, "data_dir", absoluteDir)

	go func() {
		err := server.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		if err != nil {
			logger.Error("mobile HTTP server stopped unexpectedly", "error", err)
		}
		running.serveErr <- err
	}()
	return port, nil
}

func Stop() error {
	serverState.Lock()
	defer serverState.Unlock()
	running := serverState.running
	if running == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := running.server.Shutdown(ctx); err != nil {
		closeErr := running.server.Close()
		logErr := running.logFile.Close()
		serverState.running = nil
		return errors.Join(
			fmt.Errorf("stop mobile HTTP server: %w", err),
			wrapIfError("force-close mobile HTTP server", closeErr),
			wrapIfError("close mobile application log", logErr),
		)
	}
	if err := running.logFile.Close(); err != nil {
		serverState.running = nil
		return fmt.Errorf("close mobile application log: %w", err)
	}
	serverState.running = nil
	return nil
}

func wrapIfError(action string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", action, err)
}

func Version() string {
	return version
}
