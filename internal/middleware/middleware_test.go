package middleware

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRequestIDAndCORS(t *testing.T) {
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := GetRequestID(r.Context()); got == "" {
			t.Fatal("request id missing")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if got := w.Header().Get("X-Request-ID"); got == "" {
		t.Fatal("response header missing request id")
	}

	c := CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodOptions, "/", nil)
	c.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for options, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.Header.Set("Origin", "https://attacker.example")
	c.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected cross-origin request to be denied, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.Header.Set("Origin", "http://example.com")
	c.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected same-origin request to continue, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unexpected cross-origin access header: %q", got)
	}
}

func TestRecovererAndTimeout(t *testing.T) {
	recoverer := Recoverer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	recoverer.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	timeout := Timeout(10 * time.Millisecond)
	h := timeout(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(w, req)
	if w.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d", w.Code)
	}
}

func TestLoggingAndLoggerContext(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&output, nil))
	wrapped := Logger(logger)(Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if l := GetLogger(r.Context()); l == nil {
			t.Fatal("logger missing in context")
		}
		w.WriteHeader(http.StatusOK)
	})))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	wrapped.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(output.String(), "request completed") {
		t.Fatal("expected access log entry")
	}

	ctx := context.WithValue(context.Background(), loggerKey, logger)
	if got := GetLogger(ctx); got == nil {
		t.Fatal("expected logger from context")
	}
}
