package services

import (
	"context"
	"encoding/json"
	"errors"
	"quocantran/gojob/internal/dtos"
	"quocantran/gojob/internal/models"
	"quocantran/gojob/internal/repositories"
	"strings"
)

type JobService struct {
	repository *repositories.JobRepository
}

func NewJobService(repository *repositories.JobRepository) *JobService {
	return &JobService{repository: repository}
}

func (s *JobService) CreateNewJob(ctx context.Context, jobType string, payload map[string]any) (*dtos.JobResponse, error) {
	jobType = strings.TrimSpace(jobType)

	if jobType == "" {
		return nil, errors.New("type is required")
	}

	if len(payload) == 0 {
		return nil, errors.New("payload is required")
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.New("failed to encode payload")
	}

	job := &models.Job{
		Type:    jobType,
		Payload: payloadBytes,
	}

	job, err = s.repository.CreateNewJob(ctx, job)

	if err != nil {
		return nil, errors.New("create new job failed")
	}
	var payloadMap map[string]any

	err = json.Unmarshal(job.Payload, &payloadMap) // from []bytes to map
	if err != nil {
		return nil, err
	}

	jobResponse := dtos.JobResponse{
		ID:        job.ID,
		Type:      job.Type,
		Status:    string(job.Status),
		Payload:   payloadMap,
		Attempts:  job.Attempts,
		CreatedAt: job.CreatedAt,
		UpdatedAt: job.UpdatedAt,
	}

	return &jobResponse, nil
}
