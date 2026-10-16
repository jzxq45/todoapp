//go:build !android

package task

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSQLiteRepositoryCRUDWithAndWithoutDueAt(t *testing.T) {
	ctx := context.Background()
	repository, err := NewSQLiteRepository(filepath.Join(t.TempDir(), "tasks.db"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, repository.Close()) })

	dueAt := time.Date(2026, time.October, 20, 9, 30, 0, 0, time.UTC)
	item := NewTask(CreateTaskInput{Title: "Has a due date"})
	item.DueAt = &dueAt
	require.NoError(t, repository.Create(ctx, item))

	loaded, err := repository.GetByID(ctx, item.ID)
	require.NoError(t, err)
	require.NotNil(t, loaded.DueAt)
	require.True(t, dueAt.Equal(*loaded.DueAt))

	withoutDueAt := NewTask(CreateTaskInput{Title: "No due date"})
	withoutDueAt.Done = true
	require.NoError(t, repository.Create(ctx, withoutDueAt))

	completed := true
	tasks, err := repository.List(ctx, Filter{Done: &completed})
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	require.Equal(t, withoutDueAt.ID, tasks[0].ID)
	require.Nil(t, tasks[0].DueAt)

	loaded.DueAt = nil
	loaded.Done = true
	require.NoError(t, repository.Update(ctx, loaded))
	loaded, err = repository.GetByID(ctx, item.ID)
	require.NoError(t, err)
	require.Nil(t, loaded.DueAt)
	require.True(t, loaded.Done)

	require.NoError(t, repository.Delete(ctx, withoutDueAt.ID))
	_, err = repository.GetByID(ctx, withoutDueAt.ID)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestSQLiteRepositoryMigratesLegacyTasksTable(t *testing.T) {
	ctx := context.Background()
	dsn := filepath.Join(t.TempDir(), "legacy.db")
	db, err := sql.Open("sqlite3", dsn)
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE tasks (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			done INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE INDEX idx_tasks_done ON tasks(done);
	`)
	require.NoError(t, err)
	item := NewTask(CreateTaskInput{Title: "Existing task"})
	_, err = db.Exec(
		`INSERT INTO tasks (id, title, description, done, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		item.ID.String(), item.Title, item.Description, boolToInt(item.Done),
		item.CreatedAt.Format(time.RFC3339), item.UpdatedAt.Format(time.RFC3339),
	)
	require.NoError(t, err)
	require.NoError(t, db.Close())

	repository, err := NewSQLiteRepository(dsn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, repository.Close()) })

	loaded, err := repository.GetByID(ctx, item.ID)
	require.NoError(t, err)
	require.Equal(t, item.Title, loaded.Title)
	require.Nil(t, loaded.DueAt)

	updatedTitle := "Updated legacy task"
	loaded.Title = updatedTitle
	require.NoError(t, repository.Update(ctx, loaded))
	loaded, err = repository.GetByID(ctx, item.ID)
	require.NoError(t, err)
	require.Equal(t, updatedTitle, loaded.Title)
}
