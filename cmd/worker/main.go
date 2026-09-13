package main

import (
	"context"
	"log"
	"quocantran/gojob/internal/database"
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

	workerPool := workers.NewWorkerPool(jobRepository, 1)
	workerPool.Run(ctx)
}
