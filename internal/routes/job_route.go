package routes

import (
	"net/http"
	"quocantran/gojob/internal/handlers"
)

func RegisterJobRoutes(mux *http.ServeMux, h *handlers.JobHandler){
	mux.HandleFunc("POST /jobs", h.CreateNewJob)
}