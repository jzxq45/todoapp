package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	cfg := Load()
	if cfg.Server.Host != "127.0.0.1" || cfg.Server.Port != "8080" {
		t.Fatal("expected default server config")
	}
	if cfg.Database.Driver == "" {
		t.Fatal("expected database driver")
	}
	if cfg.Log.Format == "" {
		t.Fatal("expected log format")
	}
}

func TestLoadFromEnvAndAddr(t *testing.T) {
	t.Setenv("SERVER_HOST", "127.0.0.1")
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("SERVER_READ_TIMEOUT", "15s")
	t.Setenv("SERVER_WRITE_TIMEOUT", "20s")
	t.Setenv("SERVER_IDLE_TIMEOUT", "30s")
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_DSN", "/tmp/test.db")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_FORMAT", "text")

	cfg := Load()
	if cfg.Server.Host != "127.0.0.1" || cfg.Server.Port != "9090" {
		t.Fatalf("unexpected server values: %+v", cfg.Server)
	}
	if cfg.Server.ReadTimeout != 15*time.Second || cfg.Server.WriteTimeout != 20*time.Second || cfg.Server.IdleTimeout != 30*time.Second {
		t.Fatalf("unexpected timeouts: %+v", cfg.Server)
	}
	if cfg.Database.Driver != "sqlite" || cfg.Database.DSN != "/tmp/test.db" {
		t.Fatalf("unexpected db config: %+v", cfg.Database)
	}
	if cfg.Log.Level != "debug" || cfg.Log.Format != "text" {
		t.Fatalf("unexpected log config: %+v", cfg.Log)
	}
	if cfg.Addr() != "127.0.0.1:9090" {
		t.Fatalf("unexpected addr: %s", cfg.Addr())
	}
}

func TestGetDurationEnvInvalid(t *testing.T) {
	if got := getDurationEnv("MISSING_TIMEOUT", time.Second); got != time.Second {
		t.Fatalf("expected default duration, got %v", got)
	}
	t.Setenv("MISSING_TIMEOUT", "not-a-duration")
	if got := getDurationEnv("MISSING_TIMEOUT", time.Second); got != time.Second {
		t.Fatalf("expected default after invalid duration, got %v", got)
	}
}
