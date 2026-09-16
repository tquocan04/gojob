package services

import (
	"context"
	"quocantran/gojob/internal/repositories"
	"testing"
)

// Repository methods are never reached on the validation paths under test,
// so a zero-value JobRepository (no DB pool) is sufficient.
func newTestService() *JobService {
	return NewJobService(&repositories.JobRepository{})
}

func TestCreateNewJobValidation(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	tests := []struct {
		name    string
		jobType string
		payload map[string]any
	}{
		{name: "empty type", jobType: "", payload: map[string]any{"k": "v"}},
		{name: "whitespace type", jobType: "  ", payload: map[string]any{"k": "v"}},
		{name: "empty payload", jobType: "email", payload: map[string]any{}},
		{name: "nil payload", jobType: "email", payload: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job, err := svc.CreateNewJob(ctx, tt.jobType, tt.payload)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if job != nil {
				t.Fatalf("expected nil job, got %+v", job)
			}
		})
	}
}

func TestGetJobByIdValidation(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	tests := []struct {
		name string
		id   string
	}{
		{name: "empty id", id: ""},
		{name: "not a uuid", id: "not-a-uuid"},
		{name: "malformed uuid", id: "12345678"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job, err := svc.GetJobById(ctx, tt.id)
			if err == nil {
				t.Fatal("expected invalid uuid error, got nil")
			}
			if job != nil {
				t.Fatalf("expected nil job, got %+v", job)
			}
		})
	}
}
