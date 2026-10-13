package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/jzxq45/golang-todoapp/internal/config"
	"github.com/jzxq45/golang-todoapp/internal/task"
)

func TestSetupLoggerAndParseLogLevel(t *testing.T) {
	cfg := config.LogConfig{Level: "debug", Format: "json"}
	logger := setupLogger(cfg)
	if logger == nil {
		t.Fatal("logger should not be nil")
	}
	if parseLogLevel("debug") != slog.LevelDebug {
		t.Fatal("debug log level mismatch")
	}
	if parseLogLevel("warn") != slog.LevelWarn {
		t.Fatal("warn log level mismatch")
	}
	if parseLogLevel("error") != slog.LevelError {
		t.Fatal("error log level mismatch")
	}
	if parseLogLevel("unknown") != slog.LevelInfo {
		t.Fatal("default log level mismatch")
	}
}

func TestCreateRepositoryMemoryAndSQLite(t *testing.T) {
	repo, err := createRepository(config.DatabaseConfig{Driver: "memory"}, slog.Default())
	if err != nil || repo == nil {
		t.Fatal("memory repo creation failed")
	}

	tempFile := filepath.Join(t.TempDir(), "data", "todo_test.db")
	sqlRepo, err := createRepository(config.DatabaseConfig{Driver: "sqlite", DSN: tempFile}, slog.Default())
	if err != nil {
		t.Fatal("sqlite repo creation failed")
	}
	if _, ok := sqlRepo.(*task.SQLiteRepository); !ok {
		t.Fatal("expected sqlite repository type")
	}
	if err := sqlRepo.(interface{ Close() error }).Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRouterServesFrontendAndAPI(t *testing.T) {
	router := newRouter(task.NewHandler(task.NewService(task.NewMemoryRepository())), slog.New(slog.NewTextHandler(io.Discard, nil)))

	for _, test := range []struct {
		path        string
		contentType string
		status      int
	}{
		{path: "/", contentType: "text/html; charset=utf-8", status: http.StatusOK},
		{path: "/static/app.js", contentType: "text/javascript; charset=utf-8", status: http.StatusOK},
		{path: "/static/style.css", contentType: "text/css; charset=utf-8", status: http.StatusOK},
		{path: "/some/spa/route", contentType: "text/html; charset=utf-8", status: http.StatusOK},
		{path: "/static/not-found.js", status: http.StatusNotFound},
		{path: "/api/unknown", status: http.StatusNotFound},
		{path: "/healthz", contentType: "application/json", status: http.StatusOK},
	} {
		t.Run(test.path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("status = %d, want %d", response.Code, test.status)
			}
			if test.contentType != "" && response.Header().Get("Content-Type") != test.contentType {
				t.Errorf("content type = %q, want %q", response.Header().Get("Content-Type"), test.contentType)
			}
			if response.Header().Get("Content-Security-Policy") == "" {
				t.Error("Content-Security-Policy header is missing")
			}
		})
	}
}
