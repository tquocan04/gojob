package workers

import (
	"context"
	"log"
	"quocantran/gojob/internal/models"
	"quocantran/gojob/pkg"
	"strings"
	"time"
)

const pollInterval = 60 * time.Second

type Worker struct {
	id         int
	repository JobClaimer
	processor  JobProcessor
}

func NewWorker(id int, repository JobClaimer, processor JobProcessor) *Worker {
	return &Worker{id: id, repository: repository, processor: processor}
}

func (w *Worker) Run(ctx context.Context) {
	log.Printf("Worker %d is starting ...\n", w.id)

	for {
		// Check for shutdown before every claim: once cancelled, the worker must NOT claim new jobs.
		// In-flight jobs are allowed to finish because processing happens below before the loop returns to this check.
		select {
		case <-ctx.Done():
			log.Printf("Worker %d is shutting down ...\n", w.id)
			return
		default:
		}

		job, err := w.repository.ClaimQueuedJob(ctx)
		if err != nil {
			log.Printf("Worker %d: failed to claim a job: %v\n", w.id, err)
			sleep(ctx, pollInterval)
			continue
		}

		if job == nil {
			log.Printf("Worker %d: no queued job to claim, waiting ...\n", w.id)
			sleep(ctx, pollInterval)
			continue
		}

		log.Printf("Worker %d: claimed job %s at %s (status=%s, locked_at set, attempt #%d)\n",
			w.id, job.ID, pkg.FormatVN(time.Now()), strings.ToLower(string(models.JobStatusProcessing)), job.Attempts)

		err = w.processor.Process(ctx, job)
		if err != nil {
			if job.Attempts < job.MaxAttempts {
				job.Status = models.JobStatusQueued
				job.AvailableAt = time.Now().Add(retryDelay(job.Attempts))
				log.Printf("Worker %d: job %s failed, retrying at %s (attempt %d/%d): %v\n", w.id, job.ID, pkg.FormatVN(job.AvailableAt), job.Attempts, job.MaxAttempts, err)
			} else {
				job.Status = models.JobStatusFailed
				log.Printf("Worker %d: job %s failed (attempt %d/%d): %v\n", w.id, job.ID, job.Attempts, job.MaxAttempts, err)
			}
		} else {
			job.Status = models.JobStatusCompleted
			log.Printf("Worker %d: job %s completed\n", w.id, job.ID)
		}

		log.Printf("Worker %d: processing finished, updating job %s to status=%s (attempt #%d) at %s\n",
			w.id, job.ID, job.Status, job.Attempts, pkg.FormatVN(time.Now()))

		// Use a fresh context for the final DB update so the status is persisted
		// even when the parent context has been cancelled during a graceful shutdown.
		if err := w.repository.UpdateJobStatus(context.Background(), job.ID, job.Status, job.Attempts, job.AvailableAt); err != nil {
			log.Printf("Worker %d: failed to update job status: %v\n", w.id, err)
		}
	}
}

func retryDelay(attempt int) time.Duration {
	// avoid panic with negative number case
	if attempt <= 0 {
		return 1 * time.Second
	}

	return time.Duration(1<<(attempt-1)) * time.Second
}

// sleep blocks for the given duration and returns early once the context is cancelled.
// This makes the worker's idle wait (pollInterval) interruptible so
// a graceful shutdown does not hang waiting for a sleeping worker.
func sleep(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done(): // readonly channel
		return ctx.Err()
	case <-time.After(d): // after d time, worker continue loop
		return nil
	}
}
