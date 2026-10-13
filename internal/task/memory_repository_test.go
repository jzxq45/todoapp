package task

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryRepository_Create(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	task := NewTask(CreateTaskInput{Title: "Test Task"})

	err := repo.Create(ctx, task)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, task.ID, retrieved.ID)
	assert.Equal(t, task.Title, retrieved.Title)
}

func TestMemoryRepository_GetByID_NotFound(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	_, err := repo.GetByID(ctx, uuid.New())
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMemoryRepository_List(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	task1 := NewTask(CreateTaskInput{Title: "Task 1"})
	task2 := NewTask(CreateTaskInput{Title: "Task 2"})
	task2.Done = true

	repo.Create(ctx, task1)
	repo.Create(ctx, task2)

	all, err := repo.List(ctx, Filter{})
	require.NoError(t, err)
	assert.Len(t, all, 2)

	active, err := repo.List(ctx, Filter{Done: boolPtr(false)})
	require.NoError(t, err)
	assert.Len(t, active, 1)
	assert.Equal(t, task1.ID, active[0].ID)

	completed, err := repo.List(ctx, Filter{Done: boolPtr(true)})
	require.NoError(t, err)
	assert.Len(t, completed, 1)
	assert.Equal(t, task2.ID, completed[0].ID)
}

func TestMemoryRepository_Update(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	task := NewTask(CreateTaskInput{Title: "Original"})
	repo.Create(ctx, task)

	task.Title = "Updated"
	task.Done = true
	err := repo.Update(ctx, task)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated", retrieved.Title)
	assert.True(t, retrieved.Done)
}

func TestMemoryRepository_Update_NotFound(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	task := NewTask(CreateTaskInput{Title: "Test"})
	task.ID = uuid.New()

	err := repo.Update(ctx, task)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMemoryRepository_Delete(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	task := NewTask(CreateTaskInput{Title: "Test"})
	repo.Create(ctx, task)

	err := repo.Delete(ctx, task.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, task.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMemoryRepository_Delete_NotFound(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	err := repo.Delete(ctx, uuid.New())
	assert.ErrorIs(t, err, ErrNotFound)
}
