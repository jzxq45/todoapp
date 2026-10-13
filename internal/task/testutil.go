package task

import (
	"github.com/google/uuid"
)

func boolPtr(b bool) *bool {
	return &b
}

func strPtr(s string) *string {
	return &s
}

func uuidPtr(id uuid.UUID) *uuid.UUID {
	return &id
}
