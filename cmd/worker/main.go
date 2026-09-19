package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"quocantran/gojob/internal/database"
	"quocantran/gojob/internal/processors"
	"quocantran/gojob/internal/repositories"
	"quocantran/gojob/internal/workers"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("Warning: could not load .env file: %v", err)
	}

	// Create a context cancelled by SIGINT (Ctrl+C) or SIGTERM (stop/redeploy container in Docker/Kubernetes).
	// It drives the whole graceful shutdown: workers stop claiming, let in-flight jobs finish,
	// the recovery worker stops, and finally the db pool is closed.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM) // register signals
	defer stop()                                                                           //Free up the resources of listening to the signal when the main function ends, avoiding memory leakage.

	pool, err := database.NewPostgresPool(ctx)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	// The database connection is closed LAST, only after every goroutine has
	// finished its work and returned.
	defer pool.Close()

	jobRepository := repositories.NewJobRepository(pool)
	jobProcessor := processors.NewJobProcessor()

	recoveryWorker := workers.NewRecoveryWorker(jobRepository)

	// Track the recovery worker with a WaitGroup so the app waits for it too
	// before closing the database.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		recoveryWorker.Run(ctx)
	}()

	// Run blocks until all workers exit. On cancellation they stop claiming new
	// jobs, complete the job currently being processed, then return.
	workerPool := workers.NewWorkerPool(jobRepository, 3, jobProcessor)
	workerPool.Run(ctx)

	log.Println("All processing workers exited, waiting for the recovery worker ...")

	// The recovery worker stops on the same cancelled context. Recovery is NOT
	// run during shutdown: jobs still stuck in 'processing' are left untouched
	// and will be picked up again on the next startup.
	wg.Wait()

	log.Println("Recovery worker exited, closing database connection ...")
}
