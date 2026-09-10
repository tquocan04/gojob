package repositories

import (
	"context"
	"log"
	"quocantran/gojob/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type JobRepository struct {
	db *pgxpool.Pool
}

const (
	createJobQuery = `
	insert into jobs (type, payload)
	values ($1, $2)
	returning 
	id, 
	type, 
	payload, 
	status, 
	attempts, 
	max_attempts,
	available_at,
	created_at, 
	updated_at
	`
)

func NewJobRepository(db *pgxpool.Pool) *JobRepository { // Constructor, Dependency Injection
	return &JobRepository{db: db}
}

func (r *JobRepository) CreateNewJob(ctx context.Context, job *models.Job) (*models.Job, error) {
	log.Println("===== LOGGING FROM REPOSITORY =====")
	log.Printf("Job.Type: %s\n", job.Type)
	log.Printf("Job.Payload: %s\n", job.Payload)
	row := r.db.QueryRow(ctx, createJobQuery, job.Type, job.Payload)

	var created models.Job

	// Store to db
	err := row.Scan( // Mapping with returning in sql query
		&created.ID,
		&created.Type,
		&created.Payload,
		&created.Status,
		&created.Attempts,
		&created.MaxAttempts,
		&created.AvailableAt,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &created, nil
}
