package middleware

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
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
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	wrapped := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if l := GetLogger(r.Context()); l == nil {
			t.Fatal("logger missing in context")
		}
		w.WriteHeader(http.StatusOK)
	}))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	wrapped.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	ctx := context.WithValue(context.Background(), loggerKey, logger)
	if got := GetLogger(ctx); got == nil {
		t.Fatal("expected logger from context")
	}
}
