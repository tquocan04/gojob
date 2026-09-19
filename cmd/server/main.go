package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"quocantran/gojob/internal/database"
	"quocantran/gojob/internal/handlers"
	"quocantran/gojob/internal/repositories"
	"quocantran/gojob/internal/routes"
	"quocantran/gojob/internal/services"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("Warning: could not load .env file: %v", err)
	}

	// Create a context cancelled by SIGINT (Ctrl+C) or SIGTERM so the server
	// can stop accepting new requests and drain in-flight ones gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPostgresPool(ctx)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	// Close the db pool only after the http server has fully stopped.
	defer pool.Close()

	// Jobs
	jobRepository := repositories.NewJobRepository(pool)
	jobService := services.NewJobService(jobRepository)
	jobHandler := handlers.NewJobHandler(jobService)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.HealthHandler)
	routes.RegisterJobRoutes(mux, jobHandler)

	srv := &http.Server{Addr: ":8080", Handler: mux}

	// Serve in a goroutine so the main flow can block on either a shutdown
	// signal or an unexpected server error.
	errCh := make(chan error, 1)
	go func() {
		log.Println("Server is starting ...")
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Println("Server is shutting down: refusing new requests, waiting for in-flight handlers ...")

		// Give in-flight requests a grace period to complete before force-closing.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Graceful shutdown failed: %v\n", err)
		}
		log.Println("Server shut down cleanly")
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Server error: ", err)
		}
	}
}
