package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOKAndCreated(t *testing.T) {
	w := httptest.NewRecorder()
	OK(w, map[string]string{"status": "ok"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Content-Type"); got == "" {
		t.Fatal("content type missing")
	}

	w = httptest.NewRecorder()
	Created(w, map[string]string{"id": "123"})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestNoContentAndError(t *testing.T) {
	w := httptest.NewRecorder()
	NoContent(w)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	Error(w, ErrValidation("title", "required"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if w.Body.String() == "" {
		t.Fatal("expected error payload")
	}
}

func TestHTTPStatusMapping(t *testing.T) {
	if ErrInvalidJSON.HTTPStatus() != http.StatusBadRequest {
		t.Fatal("invalid json should be 400")
	}
	if ErrNotFound.HTTPStatus() != http.StatusNotFound {
		t.Fatal("not found should be 404")
	}
	if ErrInternal.HTTPStatus() != http.StatusInternalServerError {
		t.Fatal("internal should be 500")
	}
}
