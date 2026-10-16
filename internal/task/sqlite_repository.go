//go:build !android

package task

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dsn string) (*SQLiteRepository, error) {
	if dsn != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(dsn), 0o755); err != nil {
			return nil, fmt.Errorf("create sqlite data directory: %w", err)
		}
	}
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &SQLiteRepository{db: db}, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			done INTEGER NOT NULL DEFAULT 0,
			due_at TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`)
	if err != nil {
		return err
	}

	var hasDueAt bool
	if err := db.QueryRow(`SELECT EXISTS (
		SELECT 1 FROM pragma_table_info('tasks') WHERE name = 'due_at'
	)`).Scan(&hasDueAt); err != nil {
		return err
	}
	if !hasDueAt {
		if _, err := db.Exec(`ALTER TABLE tasks ADD COLUMN due_at TEXT`); err != nil {
			return err
		}
	}

	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_tasks_done ON tasks(done);
		CREATE INDEX IF NOT EXISTS idx_tasks_due_at ON tasks(due_at);
	`)
	return err
}

func (r *SQLiteRepository) Close() error { return r.db.Close() }

func (r *SQLiteRepository) Create(ctx context.Context, task *Task) error {
	var dueAtStr *string
	if task.DueAt != nil {
		s := task.DueAt.Format(time.RFC3339)
		dueAtStr = &s
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO tasks (id, title, description, done, due_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		task.ID.String(),
		task.Title,
		task.Description,
		boolToInt(task.Done),
		dueAtStr,
		task.CreatedAt.Format(time.RFC3339),
		task.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *SQLiteRepository) GetByID(ctx context.Context, id uuid.UUID) (*Task, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, title, description, done, due_at, created_at, updated_at FROM tasks WHERE id = ?`,
		id.String(),
	)
	return scanTask(row)
}

func (r *SQLiteRepository) List(ctx context.Context, filter Filter) ([]*Task, error) {
	query := `SELECT id, title, description, done, due_at, created_at, updated_at FROM tasks`
	args := []any{}
	if filter.Done != nil {
		query += ` WHERE done = ?`
		args = append(args, boolToInt(*filter.Done))
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (r *SQLiteRepository) Update(ctx context.Context, task *Task) error {
	var dueAtStr *string
	if task.DueAt != nil {
		s := task.DueAt.Format(time.RFC3339)
		dueAtStr = &s
	}
	result, err := r.db.ExecContext(ctx,
		`UPDATE tasks SET title = ?, description = ?, done = ?, due_at = ?, updated_at = ? WHERE id = ?`,
		task.Title,
		task.Description,
		boolToInt(task.Done),
		dueAtStr,
		task.UpdatedAt.Format(time.RFC3339),
		task.ID.String(),
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SQLiteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = ?`, id.String())
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func scanTask(scanner interface{ Scan(dest ...any) error }) (*Task, error) {
	var (
		idStr, createdAtStr, updatedAtStr string
		dueAtStr                          sql.NullString
		doneInt                           int
	)
	task := &Task{}

	if err := scanner.Scan(&idStr, &task.Title, &task.Description, &doneInt, &dueAtStr, &createdAtStr, &updatedAtStr); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	task.ID = parsedID
	task.Done = doneInt == 1
	if dueAtStr.Valid {
		dueAt, err := time.Parse(time.RFC3339, dueAtStr.String)
		if err != nil {
			return nil, err
		}
		task.DueAt = &dueAt
	}
	task.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	task.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
