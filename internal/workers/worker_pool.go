package workers

import (
	"context"
	"log"
	"quocantran/gojob/internal/repositories"
	"sync"
)

type WorkerPool struct {
	repository  *repositories.JobRepository
	workerCount int
}

func NewWorkerPool(repository *repositories.JobRepository, workerCount int) *WorkerPool {
	return &WorkerPool{repository: repository, workerCount: workerCount}
}

func (p *WorkerPool) Run(ctx context.Context) {
	log.Printf("WorkerPool is starting with %d workers ...\n", p.workerCount)

	var wg sync.WaitGroup

	for i := 1; i <= p.workerCount; i++ {
		wg.Add(1)

		worker := NewWorker(i, p.repository)

		go func() {
			defer wg.Done()
			worker.Run(ctx)
		}()
	}

	wg.Wait()
}
