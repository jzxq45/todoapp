package web

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerServesAssetsAndSPA(t *testing.T) {
	handler := Handler()
	for _, test := range []struct {
		path        string
		contentType string
		wantStatus  int
		wantBody    string
	}{
		{path: "/", contentType: "text/html; charset=utf-8", wantStatus: http.StatusOK, wantBody: "Все задачи"},
		{path: "/static/app.js", contentType: "text/javascript; charset=utf-8", wantStatus: http.StatusOK, wantBody: "translations.js"},
		{path: "/static/translations.js", contentType: "text/javascript; charset=utf-8", wantStatus: http.StatusOK, wantBody: "Switch to Russian"},
		{path: "/static/ui.js", contentType: "text/javascript; charset=utf-8", wantStatus: http.StatusOK, wantBody: "function createUI"},
		{path: "/static/style.css", contentType: "text/css; charset=utf-8", wantStatus: http.StatusOK, wantBody: "--bg:"},
		{path: "/tasks/today", contentType: "text/html; charset=utf-8", wantStatus: http.StatusOK, wantBody: "Все задачи"},
		{path: "/static/missing.css", wantStatus: http.StatusNotFound},
		{path: "/api/missing", wantStatus: http.StatusNotFound},
		{path: "/api/tasks/anything", wantStatus: http.StatusNotFound},
	} {
		t.Run(test.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}

			if test.contentType != "" && response.Header().Get("Content-Type") != test.contentType {
				t.Errorf("content type = %q, want %q", response.Header().Get("Content-Type"), test.contentType)
			}
			if test.wantBody != "" && !strings.Contains(response.Body.String(), test.wantBody) {
				t.Errorf("body does not contain %q", test.wantBody)
			}
			if test.path == "/" && response.Header().Get("Cache-Control") != "no-cache" {
				t.Errorf("index Cache-Control = %q, want no-cache", response.Header().Get("Cache-Control"))
			}
			if test.path == "/static/style.css" && response.Header().Get("Cache-Control") != "no-cache" {
				t.Errorf("asset Cache-Control = %q, want no-cache", response.Header().Get("Cache-Control"))
			}
		})
	}
}

func TestEmbeddedFrontendHasRussianAndEnglish(t *testing.T) {
	script, err := fs.ReadFile(embedded, "assets/translations.js")
	if err != nil {
		t.Fatal(err)
	}
	for _, label := range []string{"Все задачи", "Активные", "Выполненные", "All tasks", "In progress", "Completed"} {
		if !strings.Contains(string(script), label) {
			t.Errorf("frontend translations do not include %q", label)
		}
	}
}

func TestHandlerRejectsUnsupportedMethods(t *testing.T) {
	response := httptest.NewRecorder()
	Handler().ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
}
