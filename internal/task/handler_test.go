package task

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer() (*httptest.ResponseRecorder, *http.Request, *Handler) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	handler := NewHandler(svc)

	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	return httptest.NewRecorder(), nil, handler
}

func TestHandler_Create(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	handler := NewHandler(svc)

	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	body := `{"title": "Test Task", "description": "Test Description"}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var task Task
	err := json.Unmarshal(w.Body.Bytes(), &task)
	require.NoError(t, err)
	assert.Equal(t, "Test Task", task.Title)
	assert.Equal(t, "Test Description", task.Description)
	assert.False(t, task.Done)
	assert.NotEqual(t, uuid.Nil, task.ID)
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	handler := NewHandler(svc)

	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	body := `{"title": "Test Task",}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_RejectsOversizedAndTrailingJSON(t *testing.T) {
	handler := NewHandler(NewService(NewMemoryRepository()))
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	largeBody := `{"title":"` + strings.Repeat("x", maxJSONBodyBytes) + `"}`
	request := httptest.NewRequest(http.MethodPost, "/api/tasks", strings.NewReader(largeBody))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, http.StatusRequestEntityTooLarge, response.Code)

	request = httptest.NewRequest(http.MethodPost, "/api/tasks", strings.NewReader(`{"title":"one"} {"title":"two"}`))
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestHandler_Create_Validation(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	handler := NewHandler(svc)

	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	// Пустое название.
	body := `{"title": ""}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_GetByID(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	handler := NewHandler(svc)

	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	// Создаём задачу перед запросом.
	created, _ := svc.Create(context.Background(), CreateTaskInput{Title: "Test"})

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/"+created.ID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var task Task
	err := json.Unmarshal(w.Body.Bytes(), &task)
	require.NoError(t, err)
	assert.Equal(t, created.ID, task.ID)
}

func TestHandler_GetByID_NotFound(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	handler := NewHandler(svc)

	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_GetByID_InvalidID(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	handler := NewHandler(svc)

	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/invalid-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_List(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	handler := NewHandler(svc)

	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	svc.Create(context.Background(), CreateTaskInput{Title: "Task 1"})
	task2, _ := svc.Create(context.Background(), CreateTaskInput{Title: "Task 2"})
	svc.Update(context.Background(), task2.ID, UpdateTaskInput{Done: boolPtr(true)})

	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var tasks []Task
	err := json.Unmarshal(w.Body.Bytes(), &tasks)
	require.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func TestHandler_List_Filter(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	handler := NewHandler(svc)

	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	svc.Create(context.Background(), CreateTaskInput{Title: "Task 1"})
	task2, _ := svc.Create(context.Background(), CreateTaskInput{Title: "Task 2"})
	svc.Update(context.Background(), task2.ID, UpdateTaskInput{Done: boolPtr(true)})

	// Фильтр выполненных задач.
	req := httptest.NewRequest(http.MethodGet, "/api/tasks?done=true", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var tasks []Task
	err := json.Unmarshal(w.Body.Bytes(), &tasks)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.True(t, tasks[0].Done)
}

func TestHandler_Update(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	handler := NewHandler(svc)

	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	created, _ := svc.Create(context.Background(), CreateTaskInput{Title: "Original"})

	body := `{"title": "Updated", "done": true}`
	req := httptest.NewRequest(http.MethodPatch, "/api/tasks/"+created.ID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var task Task
	err := json.Unmarshal(w.Body.Bytes(), &task)
	require.NoError(t, err)
	assert.Equal(t, "Updated", task.Title)
	assert.True(t, task.Done)
}

func TestHandler_UpdateCanClearDueDate(t *testing.T) {
	handler := NewHandler(NewService(NewMemoryRepository()))
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	request := httptest.NewRequest(http.MethodPost, "/api/tasks", strings.NewReader(`{"title":"Deadline","due_at":"2026-10-20T09:30:00Z"}`))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusCreated, response.Code)

	var created Task
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &created))
	require.NotNil(t, created.DueAt)

	request = httptest.NewRequest(http.MethodPatch, "/api/tasks/"+created.ID.String(), strings.NewReader(`{}`))
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &created))
	require.NotNil(t, created.DueAt)

	request = httptest.NewRequest(http.MethodPatch, "/api/tasks/"+created.ID.String(), strings.NewReader(`{"due_at":null}`))
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)

	var updated Task
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &updated))
	require.Nil(t, updated.DueAt)
}

func TestHandler_Delete(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	handler := NewHandler(svc)

	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	created, _ := svc.Create(context.Background(), CreateTaskInput{Title: "Test"})

	req := httptest.NewRequest(http.MethodDelete, "/api/tasks/"+created.ID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Проверяем удаление.
	_, err := svc.GetByID(context.Background(), created.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}
