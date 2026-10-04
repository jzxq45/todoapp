package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/google/uuid"
)

type JSONRepository struct {
	mu       sync.RWMutex
	filePath string
	tasks    map[uuid.UUID]Task
}

func NewJSONRepository(filePath string) (*JSONRepository, error) {
	if filePath == "" {
		return nil, errors.New("JSON repository path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0o700); err != nil {
		return nil, fmt.Errorf("create JSON repository directory: %w", err)
	}

	repository := &JSONRepository{
		filePath: filePath,
		tasks:    make(map[uuid.UUID]Task),
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("read JSON repository: %w", err)
		}
		if err := repository.save(repository.tasks); err != nil {
			return nil, fmt.Errorf("initialize JSON repository: %w", err)
		}
		return repository, nil
	}

	var tasks []Task
	if err := json.Unmarshal(content, &tasks); err != nil {
		return nil, fmt.Errorf("decode JSON repository: %w", err)
	}
	for _, item := range tasks {
		if item.ID == uuid.Nil {
			return nil, errors.New("decode JSON repository: task has empty ID")
		}
		if _, exists := repository.tasks[item.ID]; exists {
			return nil, fmt.Errorf("decode JSON repository: duplicate task ID %s", item.ID)
		}
		if err := item.Validate(); err != nil {
			return nil, fmt.Errorf("decode JSON repository task %s: %w", item.ID, err)
		}
		repository.tasks[item.ID] = item
	}
	return repository, nil
}

func (r *JSONRepository) Create(ctx context.Context, item *Task) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tasks[item.ID]; exists {
		return fmt.Errorf("create task %s: duplicate ID", item.ID)
	}
	next := cloneTasks(r.tasks)
	next[item.ID] = *item
	if err := r.save(next); err != nil {
		return fmt.Errorf("persist created task: %w", err)
	}
	r.tasks = next
	return nil
}

func (r *JSONRepository) GetByID(ctx context.Context, id uuid.UUID) (*Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, exists := r.tasks[id]
	if !exists {
		return nil, ErrNotFound
	}
	copy := item
	return &copy, nil
}

func (r *JSONRepository) List(ctx context.Context, filter Filter) ([]*Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]Task, 0, len(r.tasks))
	for _, item := range r.tasks {
		if filter.Done == nil || item.Done == *filter.Done {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].ID.String() < items[j].ID.String()
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	result := make([]*Task, 0, len(items))
	for i := range items {
		item := items[i]
		result = append(result, &item)
	}
	return result, nil
}

func (r *JSONRepository) Update(ctx context.Context, item *Task) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tasks[item.ID]; !exists {
		return ErrNotFound
	}
	next := cloneTasks(r.tasks)
	next[item.ID] = *item
	if err := r.save(next); err != nil {
		return fmt.Errorf("persist updated task: %w", err)
	}
	r.tasks = next
	return nil
}

func (r *JSONRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tasks[id]; !exists {
		return ErrNotFound
	}
	next := cloneTasks(r.tasks)
	delete(next, id)
	if err := r.save(next); err != nil {
		return fmt.Errorf("persist deleted task: %w", err)
	}
	r.tasks = next
	return nil
}

func (r *JSONRepository) save(tasks map[uuid.UUID]Task) error {
	items := make([]Task, 0, len(tasks))
	for _, item := range tasks {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].ID.String() < items[j].ID.String()
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	content, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("encode tasks: %w", err)
	}

	temp, err := os.CreateTemp(filepath.Dir(r.filePath), ".tasks-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary repository file: %w", err)
	}
	tempName := temp.Name()
	defer os.Remove(tempName)

	if err := temp.Chmod(0o600); err != nil {
		_ = temp.Close()
		return fmt.Errorf("set temporary repository permissions: %w", err)
	}
	if _, err := temp.Write(content); err != nil {
		_ = temp.Close()
		return fmt.Errorf("write temporary repository file: %w", err)
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return fmt.Errorf("sync temporary repository file: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close temporary repository file: %w", err)
	}
	if err := os.Rename(tempName, r.filePath); err != nil {
		return fmt.Errorf("replace repository file: %w", err)
	}
	return nil
}

func cloneTasks(source map[uuid.UUID]Task) map[uuid.UUID]Task {
	clone := make(map[uuid.UUID]Task, len(source)+1)
	for id, item := range source {
		clone[id] = item
	}
	return clone
}
