package models

import (
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	JobStatusQueued     JobStatus = "queued"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

type Job struct {
	ID          uuid.UUID  `json:"id"`
	Type        string     `json:"type"`
	Payload     []byte     `json:"payload"` //map JSONB in database layer
	Status      JobStatus  `json:"status"`
	Attempts    int        `json:"attempts"`
	MaxAttempts int        `json:"max_attempts"`
	AvailableAt time.Time  `json:"available_at"`
	LockedAt    *time.Time `json:"locked_at,omitempty"` // null
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
