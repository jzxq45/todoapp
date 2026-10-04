package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jzxq45/golang-todoapp/internal/config"
	"github.com/jzxq45/golang-todoapp/internal/httpapp"
	"github.com/jzxq45/golang-todoapp/internal/task"
)

var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg := config.Load()
	logger := setupLogger(cfg.Log)
	slog.SetDefault(logger)

	repo, err := createRepository(cfg.Database, logger)
	if err != nil {
		return fmt.Errorf("create repository: %w", err)
	}
	if closer, ok := repo.(interface{ Close() error }); ok {
		defer func() {
			if err := closer.Close(); err != nil {
				logger.Error("failed to close repository", "error", err)
			}
		}()
	}

	router := newRouter(task.NewHandler(task.NewService(repo)), logger)
	server := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	listener, err := net.Listen("tcp", cfg.Addr())
	if err != nil {
		return fmt.Errorf("cannot listen on %s (check that the port is available and the address is valid): %w", cfg.Addr(), err)
	}

	logger.Info("server ready",
		"version", version,
		"address", cfg.Addr(),
		"url", fmt.Sprintf("http://localhost:%s", cfg.Server.Port),
		"mode", cfg.Database.Driver,
		"data_path", cfg.Database.DSN,
	)

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(listener)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}
		return nil
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	}
}

func newRouter(handler *task.Handler, logger *slog.Logger) *chi.Mux {
	return httpapp.New(handler, logger)
}

func setupLogger(cfg config.LogConfig) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLogLevel(cfg.Level)}
	if cfg.Format == "json" {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func createRepository(cfg config.DatabaseConfig, logger *slog.Logger) (task.Repository, error) {
	switch cfg.Driver {
	case "sqlite":
		logger.Info("using SQLite repository", "dsn", cfg.DSN)
		return task.NewSQLiteRepository(cfg.DSN)
	case "memory":
		logger.Info("using in-memory repository")
		return task.NewMemoryRepository(), nil
	default:
		return nil, fmt.Errorf("unsupported database driver %q", cfg.Driver)
	}
}
