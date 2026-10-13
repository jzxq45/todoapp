package task

import (
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type Task struct {
	ID        uuid.UUID  `json:"id"`
	Title     string     `json:"title"`
	Description string   `json:"description"`
	Done      bool       `json:"done"`
	DueAt     *time.Time `json:"due_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CreateTaskInput struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	DueAt       *time.Time `json:"due_at,omitempty"`
}

type UpdateTaskInput struct {
	Title       *string     `json:"title,omitempty"`
	Description *string     `json:"description,omitempty"`
	Done        *bool       `json:"done,omitempty"`
	DueAt       *time.Time  `json:"due_at,omitempty"`
}

type Filter struct {
	Done *bool
}

func NewTask(input CreateTaskInput) *Task {
	now := time.Now().UTC()
	return &Task{
		ID:          uuid.New(),
		Title:       input.Title,
		Description: input.Description,
		Done:        false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (t *Task) Update(input UpdateTaskInput) {
	if input.Title != nil {
		t.Title = *input.Title
	}
	if input.Description != nil {
		t.Description = *input.Description
	}
	if input.Done != nil {
		t.Done = *input.Done
	}

	now := time.Now().UTC()
	if !now.After(t.UpdatedAt) {
		now = t.UpdatedAt.Add(time.Second)
	}
	t.UpdatedAt = now
}

func (t *Task) Validate() error {
	if t.Title == "" {
		return ErrTitleRequired
	}
	if utf8.RuneCountInString(t.Title) > 200 {
		return ErrTitleTooLong
	}
	return nil
}
