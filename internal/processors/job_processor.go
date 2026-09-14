package processors

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"quocantran/gojob/internal/models"
	"strings"
	"time"
)

type JobProcessor struct{}

func NewJobProcessor() *JobProcessor {
	return &JobProcessor{}
}

func (p *JobProcessor) Process(ctx context.Context, job *models.Job) error {
	log.Printf("Processing job %s type=%s payload=%s\n", job.ID, job.Type, job.Payload)

	jobType := strings.ToLower(job.Type)

	switch jobType {
	case "email":
		return p.processEmail(job)
	case "report":
		return p.processReport(job)
	case "image":
		return p.processImage(job)
	default:
		return fmt.Errorf("unsupported job type %q", job.Type)
	}
}

func (p *JobProcessor) processEmail(job *models.Job) error {
	return p.processWithDelay(job, 1*time.Second, "sending email")
}

func (p *JobProcessor) processReport(job *models.Job) error {
	return p.processWithDelay(job, 3*time.Second, "generating report")
}

func (p *JobProcessor) processImage(job *models.Job) error {
	return p.processWithDelay(job, 3*time.Second, "generating report")
}

func (p *JobProcessor) processWithDelay(job *models.Job, delay time.Duration, action string) error {
	log.Printf("Job %s: %s ...\n", job.ID, action)
	time.Sleep(delay)

	// random
	if rand.Intn(100) < 80 {
		log.Printf("Job %s: %s succeeded\n", job.ID, action)
		return nil
	}

	log.Printf("Job %s: %s failed\n", job.ID, action)
	return errors.New(action + " failed")
}
