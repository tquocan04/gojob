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
		WriteError(w, http.StatusBadRequest, "failed to parse request body")
		return
	}

	job, err := h.service.CreateNewJob(r.Context(), req.Type, req.Payload)

	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, "create successful", job)
}

func (h *JobHandler) GetJobById(w http.ResponseWriter, r *http.Request) {
	jobId := r.PathValue("id")

	job, err := h.service.GetJobById(r.Context(), jobId)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	} else if job == nil {
		WriteError(w, http.StatusNotFound, "no jobs found")
		return
	}

	WriteJSON(w, http.StatusOK, "successful", job)
}
