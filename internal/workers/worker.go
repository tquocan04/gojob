package workers

import (
	"context"
	"log"
	"quocantran/gojob/internal/models"
	"quocantran/gojob/internal/processors"
	"quocantran/gojob/internal/repositories"
	"time"
)

const pollInterval = 60 * time.Second

type Worker struct {
	id         int
	repository *repositories.JobRepository
	processor  *processors.JobProcessor
}

func NewWorker(id int, repository *repositories.JobRepository, processor *processors.JobProcessor) *Worker {
	return &Worker{id: id, repository: repository, processor: processor}
}

func (w *Worker) Run(ctx context.Context) {
	log.Printf("Worker %d is starting ...\n", w.id)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d is shutting down ...\n", w.id)
			return
		default:
		}

		job, err := w.repository.ClaimQueuedJob(ctx)
		if err != nil {
			log.Printf("Worker %d: failed to claim a job: %v\n", w.id, err)
			time.Sleep(pollInterval)
			continue
		}

		if job == nil {
			log.Printf("Worker %d: no queued job to claim, waiting ...\n", w.id)
			time.Sleep(pollInterval)
			continue
		}

		err = w.processor.Process(ctx, job)
		if err != nil {
			if job.Attempts < job.MaxAttempts {
				job.Status = models.JobStatusQueued
				job.AvailableAt = time.Now().Add(retryDelay(job.Attempts))
				log.Printf("Worker %d: job %s failed, retrying at %s (attempt %d/%d): %v\n", w.id, job.ID, job.AvailableAt.Format("15:04:05"), job.Attempts, job.MaxAttempts, err)
			} else {
				job.Status = models.JobStatusFailed
				log.Printf("Worker %d: job %s failed (attempt %d/%d): %v\n", w.id, job.ID, job.Attempts, job.MaxAttempts, err)
			}
		} else {
			job.Status = models.JobStatusCompleted
			log.Printf("Worker %d: job %s completed\n", w.id, job.ID)
		}

		err = w.repository.UpdateJobStatus(ctx, job.ID, job.Status, job.Attempts, job.AvailableAt)
		if err != nil {
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
