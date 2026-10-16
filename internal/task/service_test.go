package task

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	task, err := svc.Create(ctx, CreateTaskInput{Title: "Test Task", Description: "Desc"})
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, task.ID)
	assert.Equal(t, "Test Task", task.Title)
	assert.Equal(t, "Desc", task.Description)
	assert.False(t, task.Done)
	assert.False(t, task.CreatedAt.IsZero())
	assert.False(t, task.UpdatedAt.IsZero())
}

func TestService_Create_Validation(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	// Пустое название.
	_, err := svc.Create(ctx, CreateTaskInput{Title: ""})
	assert.ErrorIs(t, err, ErrTitleRequired)
	_, err = svc.Create(ctx, CreateTaskInput{Title: " \t\n "})
	assert.ErrorIs(t, err, ErrTitleRequired)

	// Слишком длинное название.
	longTitle := string(make([]byte, 201))
	_, err = svc.Create(ctx, CreateTaskInput{Title: longTitle})
	assert.ErrorIs(t, err, ErrTitleTooLong)

	unicodeTitle := strings.Repeat("界", 200)
	_, err = svc.Create(ctx, CreateTaskInput{Title: unicodeTitle})
	assert.NoError(t, err, "200 Unicode characters should be accepted")
	_, err = svc.Create(ctx, CreateTaskInput{Title: unicodeTitle + "界"})
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

func TestService_GetByID(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	created, err := svc.Create(ctx, CreateTaskInput{Title: "Test"})
	require.NoError(t, err)

	task, err := svc.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, task.ID)
}

func TestService_GetByID_NotFound(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	_, err := svc.GetByID(ctx, uuid.New())
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestService_List(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	svc.Create(ctx, CreateTaskInput{Title: "Task 1"})
	task2, _ := svc.Create(ctx, CreateTaskInput{Title: "Task 2"})
	svc.Update(ctx, task2.ID, UpdateTaskInput{Done: boolPtr(true)})

	all, err := svc.List(ctx, Filter{})
	require.NoError(t, err)
	assert.Len(t, all, 2)

	active, err := svc.List(ctx, Filter{Done: boolPtr(false)})
	require.NoError(t, err)
	assert.Len(t, active, 1)

	completed, err := svc.List(ctx, Filter{Done: boolPtr(true)})
	require.NoError(t, err)
	assert.Len(t, completed, 1)
}

func TestService_Update(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	created, err := svc.Create(ctx, CreateTaskInput{Title: "Original"})
	require.NoError(t, err)

	updated, err := svc.Update(ctx, created.ID, UpdateTaskInput{
		Title:       strPtr("Updated"),
		Description: strPtr("New desc"),
		Done:        boolPtr(true),
	})
	require.NoError(t, err)

	assert.Equal(t, "Updated", updated.Title)
	assert.Equal(t, "New desc", updated.Description)
	assert.True(t, updated.Done)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt))
}

func TestService_Update_Validation(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	created, _ := svc.Create(ctx, CreateTaskInput{Title: "Original"})

	// Пустое название.
	_, err := svc.Update(ctx, created.ID, UpdateTaskInput{Title: strPtr("")})
	assert.ErrorIs(t, err, ErrTitleRequired)

	// Слишком длинное название.
	longTitle := string(make([]byte, 201))
	_, err = svc.Update(ctx, created.ID, UpdateTaskInput{Title: strPtr(longTitle)})
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

func TestService_Update_NotFound(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	_, err := svc.Update(ctx, uuid.New(), UpdateTaskInput{Title: strPtr("Test")})
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestService_Delete(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	created, _ := svc.Create(ctx, CreateTaskInput{Title: "Test"})

	err := svc.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = svc.GetByID(ctx, created.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestService_Delete_NotFound(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	err := svc.Delete(ctx, uuid.New())
	assert.ErrorIs(t, err, ErrNotFound)
}
