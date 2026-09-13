package repositories

import (
	"context"
	"errors"
	"log"
	"quocantran/gojob/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

	getJobByIdQuery = `
	select id, status
	from jobs
	where id = $1
	`

	getQueuedJobQuery = `
	select
		id,
		type,
		payload,
		status,
		attempts,
		max_attempts,
		available_at,
		created_at,
		updated_at
	from jobs
	where status = 'queued'
	order by created_at
	limit 1
	`

	updateJobStatusQuery = `
	update jobs
	set status = $2, updated_at = now()
	where id = $1
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

func (r *JobRepository) GetJobById(ctx context.Context, id string) (*models.Job, error) {
	row := r.db.QueryRow(ctx, getJobByIdQuery, id)

	var job models.Job

	err := row.Scan(&job.ID, &job.Status)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (r *JobRepository) GetQueuedJob(ctx context.Context) (*models.Job, error) {
	row := r.db.QueryRow(ctx, getQueuedJobQuery)

	var job models.Job

	err := row.Scan(
		&job.ID,
		&job.Type,
		&job.Payload,
		&job.Status,
		&job.Attempts,
		&job.MaxAttempts,
		&job.AvailableAt,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (r *JobRepository) UpdateJobStatus(ctx context.Context, id uuid.UUID, status models.JobStatus) error {
	_, err := r.db.Exec(ctx, updateJobStatusQuery, id, status)
	return err
}
