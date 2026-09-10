package services

import (
	"context"
	"encoding/json"
	"errors"
	"quocantran/gojob/internal/dtos"
	"quocantran/gojob/internal/models"
	"quocantran/gojob/internal/repositories"
	"strings"
	"github.com/google/uuid"
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
		ID:     job.ID,
		Status: string(job.Status),
	}

	return &jobResponse, nil
}

func (s *JobService) GetJobById(ctx context.Context, id string) (*dtos.JobResponse, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid uuid format")
	}

	job, err := s.repository.GetJobById(ctx, id)
	if err != nil {
		return nil, err
	}

	if job == nil {
		return nil, nil
	}

	jobDto := dtos.JobResponse{
		ID:     job.ID,
		Status: string(job.Status),
	}

	return &jobDto, nil
}
