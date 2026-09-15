package main

import (
	"context"
	"log"
	"quocantran/gojob/internal/database"
	"quocantran/gojob/internal/processors"
	"quocantran/gojob/internal/repositories"
	"quocantran/gojob/internal/workers"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Env not found")
	}

	ctx := context.Background()

	pool, err := database.NewPostgresPool(ctx)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	defer pool.Close()

	jobRepository := repositories.NewJobRepository(pool)
	jobProcessor := processors.NewJobProcessor()

	recoveryWorker := workers.NewRecoveryWorker(jobRepository)
	go recoveryWorker.Run(ctx)

	workerPool := workers.NewWorkerPool(jobRepository, 3, jobProcessor)
	workerPool.Run(ctx)
}
