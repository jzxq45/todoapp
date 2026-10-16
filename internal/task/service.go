package task

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrTitleRequired = errors.New("title is required")
	ErrTitleTooLong  = errors.New("title must be at most 200 characters")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CreateTaskInput) (*Task, error) {
	if err := validateTitle(input.Title); err != nil {
		return nil, err
	}

	task := NewTask(input)
	task.DueAt = input.DueAt
	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Task, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, filter Filter) ([]*Task, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateTaskInput) (*Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		if err := validateTitle(*input.Title); err != nil {
			return nil, err
		}
	}

	task.Update(input)
	if input.DueAt.Set {
		task.DueAt = input.DueAt.Value
	}
	if err := s.repo.Update(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
