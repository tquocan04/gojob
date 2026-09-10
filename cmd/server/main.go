package main

import (
	"context"
	"log"
	"net/http"
	"quocantran/gojob/internal/database"
	"quocantran/gojob/internal/handlers"

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

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.HealthHandler)

	log.Println("Server is starting ...")
	err = http.ListenAndServe(":8080", mux)

	if err != nil {
		log.Fatal("Server error. Shutting down ...") // Print and Exit
	}
}
