package handlers

import (
	"encoding/json"
	"net/http"
	"quocantran/gojob/internal/dtos"
	"quocantran/gojob/internal/services"
)

type JobHandler struct {
	service *services.JobService
}

func NewJobHandler(service *services.JobService) *JobHandler {
	return &JobHandler{service: service}
}

func (h *JobHandler) CreateNewJob(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreateJobRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http.Error(w, "failed", http.StatusInternalServerError)
		return
	}

	job, err := h.service.CreateNewJob(r.Context(), req.Type, req.Payload)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]any{
		"status": "create successful",
		"data":   job,
	}

	data, err := json.Marshal(response)

	if err != nil {
		http.Error(w, "json encoding failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

func (h *JobHandler) GetJobById(w http.ResponseWriter, r *http.Request) {
	jobId := r.PathValue("id")

	job, err := h.service.GetJobById(r.Context(), jobId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if job == nil {
		http.Error(w, "no jobs found", http.StatusNotFound)
		return
	}

	response := map[string]any{
		"data":    job,
		"message": "successful",
	}

	data, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "json encoding failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
