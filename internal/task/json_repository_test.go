package task

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestJSONRepositoryPersistsAndFiltersTasks(t *testing.T) {
	ctx := context.Background()
	filePath := filepath.Join(t.TempDir(), "nested", "tasks.json")
	repository, err := NewJSONRepository(filePath)
	require.NoError(t, err)

	first := NewTask(CreateTaskInput{Title: "Первое"})
	require.NoError(t, repository.Create(ctx, first))
	second := NewTask(CreateTaskInput{Title: "Второе"})
	second.Done = true
	require.NoError(t, repository.Create(ctx, second))

	reopened, err := NewJSONRepository(filePath)
	require.NoError(t, err)
	all, err := reopened.List(ctx, Filter{})
	require.NoError(t, err)
	require.Len(t, all, 2)
	require.Equal(t, second.ID, all[0].ID)

	done := true
	completed, err := reopened.List(ctx, Filter{Done: &done})
	require.NoError(t, err)
	require.Len(t, completed, 1)
	require.Equal(t, second.ID, completed[0].ID)

	require.NoError(t, reopened.Delete(ctx, first.ID))
	require.NoError(t, reopened.Delete(ctx, second.ID))
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.JSONEq(t, "[]", string(content))
}

func TestJSONRepositoryReturnsCopiesAndNotFound(t *testing.T) {
	ctx := context.Background()
	repository, err := NewJSONRepository(filepath.Join(t.TempDir(), "tasks.json"))
	require.NoError(t, err)
	item := NewTask(CreateTaskInput{Title: "До"})
	require.NoError(t, repository.Create(ctx, item))

	loaded, err := repository.GetByID(ctx, item.ID)
	require.NoError(t, err)
	loaded.Title = "Не сохранять напрямую"
	loaded, err = repository.GetByID(ctx, item.ID)
	require.NoError(t, err)
	require.Equal(t, "До", loaded.Title)

	require.ErrorIs(t, repository.Delete(ctx, uuid.New()), ErrNotFound)
}

func TestJSONRepositoryPersistsUpdates(t *testing.T) {
	ctx := context.Background()
	filePath := filepath.Join(t.TempDir(), "tasks.json")
	repository, err := NewJSONRepository(filePath)
	require.NoError(t, err)
	item := NewTask(CreateTaskInput{Title: "До изменения"})
	require.NoError(t, repository.Create(ctx, item))

	item.Title = "После изменения"
	item.Done = true
	require.NoError(t, repository.Update(ctx, item))

	reopened, err := NewJSONRepository(filePath)
	require.NoError(t, err)
	persisted, err := reopened.GetByID(ctx, item.ID)
	require.NoError(t, err)
	require.Equal(t, "После изменения", persisted.Title)
	require.True(t, persisted.Done)
}

func TestJSONRepositoryRejectsCorruptedData(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "tasks.json")
	require.NoError(t, os.WriteFile(filePath, []byte("{"), 0o600))
	_, err := NewJSONRepository(filePath)
	require.ErrorContains(t, err, "decode JSON repository")
}
