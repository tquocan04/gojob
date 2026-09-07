package main

import (
	"log"
	"net/http"
	"quocantran/gojob/internal/handlers"
)

func main() {
	http.HandleFunc("/health", handlers.HealthHandler)
	log.Println("Server is starting ...")
	http.ListenAndServe(":8080", nil)
}
