package mobile

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestEmbeddedServerPersistsTasksAcrossRestarts(t *testing.T) {
	dataDir := t.TempDir()
	t.Cleanup(func() {
		if err := Stop(); err != nil {
			t.Errorf("stop mobile server: %v", err)
		}
	})
	if Version() == "" {
		t.Fatal("version must not be empty")
	}
	if _, err := Start(""); err == nil {
		t.Fatal("starting with an empty data directory should fail")
	}

	port, err := Start(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	reusedPort, err := Start(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	if reusedPort != port {
		t.Fatalf("reused port = %d, want %d", reusedPort, port)
	}
	if _, err := Start(t.TempDir()); err == nil {
		t.Fatal("starting with a second data directory should fail")
	}

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	response := mustRequest(t, http.MethodGet, baseURL+"/healthz", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("health status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	drainAndClose(t, response.Body)

	response = mustRequest(t, http.MethodGet, baseURL+"/", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("web UI status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	html, err := io.ReadAll(response.Body)
	drainAndClose(t, response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(string(html)), "<!doctype html") {
		t.Fatal("web UI response does not contain the embedded HTML document")
	}

	response = mustRequest(t, http.MethodGet, baseURL+"/static/app.js", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("embedded JavaScript status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	drainAndClose(t, response.Body)

	response = mustRequest(t, http.MethodPost, baseURL+"/api/tasks", `{"title":"Offline task"}`)
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", response.StatusCode, http.StatusCreated)
	}
	response.Body.Close()
	if err := Stop(); err != nil {
		t.Fatal(err)
	}

	port, err = Start(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	response = mustRequest(t, http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/api/tasks", port), "")
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	content, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	var tasks []struct {
		Title string `json:"title"`
	}
	if err := json.Unmarshal(content, &tasks); err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].Title != "Offline task" {
		t.Fatalf("persisted tasks = %s", strings.TrimSpace(string(content)))
	}
}

func mustRequest(t *testing.T, method, url, body string) *http.Response {
	t.Helper()
	request, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func drainAndClose(t *testing.T, body io.ReadCloser) {
	t.Helper()
	if _, err := io.Copy(io.Discard, body); err != nil {
		t.Errorf("read response body: %v", err)
	}
	if err := body.Close(); err != nil {
		t.Errorf("close response body: %v", err)
	}
}
