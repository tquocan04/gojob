package dtos

import (
	"github.com/google/uuid"
)

type CreateJobRequest struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
}

type JobResponse struct {
	ID        uuid.UUID      `json:"id"`
	Type      string         `json:"type"`
	Payload   map[string]any `json:"payload"`
	Status    string         `json:"status"`
	Attempts  int            `json:"attempts"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
}
