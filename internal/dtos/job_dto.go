package dtos

import (
	"github.com/google/uuid"
)

type CreateJobRequest struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
}

type JobResponse struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}
