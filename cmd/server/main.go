package main

import (
	"log"
	"net/http"
	"quocantran/gojob/internal/handlers"
)

func main() {
	http.HandleFunc("/health", handlers.HealthHandler)
	log.Println("Server is starting ...")
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal("Server error. Shutting down ...") // Print and Exit
	}
}
