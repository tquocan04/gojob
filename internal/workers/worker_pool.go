package workers

import (
	"context"
	"log"
	"quocantran/gojob/internal/processors"
	"quocantran/gojob/internal/repositories"
	"sync"
)

type WorkerPool struct {
	repository  *repositories.JobRepository
	workerCount int
	processor   *processors.JobProcessor
}

func NewWorkerPool(repository *repositories.JobRepository, workerCount int, processor *processors.JobProcessor) *WorkerPool {
	return &WorkerPool{repository: repository, workerCount: workerCount, processor: processor}
}

func (p *WorkerPool) Run(ctx context.Context) {
	log.Printf("WorkerPool is starting with %d workers ...\n", p.workerCount)

	var wg sync.WaitGroup

	for i := 1; i <= p.workerCount; i++ {
		wg.Add(1)

		worker := NewWorker(i, p.repository, p.processor)

		go func() {
			defer wg.Done()
			worker.Run(ctx)
		}()
	}

	wg.Wait()
}
