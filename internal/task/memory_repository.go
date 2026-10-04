package task

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

type MemoryRepository struct {
	mu    sync.RWMutex
	tasks map[uuid.UUID]*Task
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{tasks: make(map[uuid.UUID]*Task)}
}

func (r *MemoryRepository) Create(ctx context.Context, task *Task) error {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[task.ID] = task
	return nil
}

func (r *MemoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*Task, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()
	task, ok := r.tasks[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneTask(task), nil
}

func (r *MemoryRepository) List(ctx context.Context, filter Filter) ([]*Task, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		if filter.Done != nil && task.Done != *filter.Done {
			continue
		}
		result = append(result, cloneTask(task))
	}
	return result, nil
}

func (r *MemoryRepository) Update(ctx context.Context, task *Task) error {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tasks[task.ID]; !ok {
		return ErrNotFound
	}
	r.tasks[task.ID] = task
	return nil
}

func (r *MemoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tasks[id]; !ok {
		return ErrNotFound
	}
	delete(r.tasks, id)
	return nil
}

func cloneTask(task *Task) *Task {
	if task == nil {
		return nil
	}
	copy := *task
	return &copy
}
