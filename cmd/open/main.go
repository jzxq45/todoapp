package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

func main() {
	if err := openApp(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func openApp() error {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	if _, err := strconv.Atoi(port); err != nil {
		return fmt.Errorf("invalid SERVER_PORT %q: %w", port, err)
	}
	url := "http://localhost:" + port
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	binary, err := os.CreateTemp("", "todoapp-open-*.exe")
	if err != nil {
		return fmt.Errorf("create temporary server binary: %w", err)
	}
	binaryPath := binary.Name()
	if err := binary.Close(); err != nil {
		_ = os.Remove(binaryPath)
		return fmt.Errorf("close temporary server binary: %w", err)
	}
	defer os.Remove(binaryPath)

	build := exec.CommandContext(ctx, "go", "build", "-o", filepath.Clean(binaryPath), "./cmd/server")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("build server: %w", err)
	}
	server := exec.Command(binaryPath)
	server.Stdout = os.Stdout
	server.Stderr = os.Stderr
	if err := server.Start(); err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	done := make(chan error, 1)
	go func() { done <- server.Wait() }()

	client := &http.Client{Timeout: time.Second}
	deadline := time.NewTimer(30 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case err := <-done:
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("server exited before becoming healthy: %w", err)
		case <-ctx.Done():
			_ = server.Process.Kill()
			<-done
			return nil
		case <-deadline.C:
			_ = server.Process.Kill()
			<-done
			return fmt.Errorf("server did not become ready within 30 seconds; open %s after checking the logs", url)
		case <-ticker.C:
			response, err := client.Get(url + "/healthz")
			if err != nil {
				continue
			}
			_ = response.Body.Close()
			if response.StatusCode != http.StatusOK {
				continue
			}
			fmt.Printf("Opening %s\n", url)
			if err := launchBrowser(url); err != nil {
				fmt.Fprintf(os.Stderr, "Could not open a browser: %v\nOpen %s manually.\n", err, url)
			}
			select {
			case <-ctx.Done():
				_ = server.Process.Kill()
				<-done
				return nil
			case err := <-done:
				if errors.Is(err, context.Canceled) {
					return nil
				}
				if err != nil {
					return fmt.Errorf("server stopped: %w", err)
				}
				return nil
			}
		}
	}
}

func launchBrowser(url string) error {
	var command string
	var args []string
	switch runtime.GOOS {
	case "windows":
		command, args = "rundll32", []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		command, args = "open", []string{url}
	default:
		command, args = "xdg-open", []string{url}
	}
	return exec.Command(command, args...).Start()
}
