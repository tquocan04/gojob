package workers

import (
	"context"
	"log"
	"quocantran/gojob/internal/models"
	"quocantran/gojob/internal/repositories"
	"time"
)

const pollInterval = 60 * time.Second

type Worker struct {
	id         int
	repository *repositories.JobRepository
}

func NewWorker(id int, repository *repositories.JobRepository) *Worker {
	return &Worker{id: id, repository: repository}
}

func (w *Worker) Run(ctx context.Context) {
	log.Printf("Worker %d is starting ...\n", w.id)

	for {
		job, err := w.repository.GetQueuedJob(ctx)
		if err != nil {
			log.Printf("Worker %d: failed to get a job: %v\n", w.id, err)
			time.Sleep(pollInterval)
			continue
		}

		if job == nil {
			log.Printf("Worker %d: no queued job, waiting ...\n", w.id)
			time.Sleep(pollInterval)
			continue
		}

		w.processJob(ctx, job)

		if job.Status == models.JobStatusCompleted {
			log.Printf("Worker %d: job %s completed\n", w.id, job.ID)
		} else {
			log.Printf("Worker %d: job %s failed\n", w.id, job.ID)
		}

		err = w.repository.UpdateJobStatus(ctx, job.ID, job.Status)
		if err != nil {
			log.Printf("Worker %d: failed to update job status: %v\n", w.id, err)
		}
	}
}

func (w *Worker) processJob(ctx context.Context, job *models.Job) {
	log.Printf("Worker %d: processing job %s type=%s payload=%s\n", w.id, job.ID, job.Type, job.Payload)

	// TODO: real processing logic will be added in the next phases.
	job.Status = models.JobStatusCompleted
}
