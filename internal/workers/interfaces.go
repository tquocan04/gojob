package workers

import (
	"context"
	"quocantran/gojob/internal/models"
	"time"

	"github.com/google/uuid"
)

// JobClaimer is the minimal repository surface the worker needs to claim and
// update jobs. The concrete repositories.JobRepository satisfies it.
type JobClaimer interface {
	ClaimQueuedJob(ctx context.Context) (*models.Job, error)
	UpdateJobStatus(ctx context.Context, id uuid.UUID, status models.JobStatus, attempts int, availableAt time.Time) error
}

// JobProcessor processes a single claimed job. The concrete
// processors.JobProcessor satisfies it.
type JobProcessor interface {
	Process(ctx context.Context, job *models.Job) error
}
