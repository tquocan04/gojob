package repositories

import (
	"context"
	"errors"
	"log"
	"quocantran/gojob/internal/models"
	"time"

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

	claimQueuedJobQuery = `
	select
		id,
		type,
		payload,
		status,
		attempts,
		max_attempts,
		available_at,
		locked_at,
		created_at,
		updated_at
	from jobs
	where status = 'queued'
		and available_at <= now()
	order by created_at
	limit 1
	for update skip locked
	`

claimQueuedJobUpdateQuery = `
	update jobs
	set status = 'processing',
		attempts = attempts + 1,
		locked_at = now(),
		updated_at = now()
	where id = $1
	returning attempts
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
		and available_at <= now()
	order by created_at
	limit 1
	`

	updateJobStatusQuery = `
	update jobs
	set status = $2, attempts = $3, available_at = $4, updated_at = now()
	where id = $1
	`

	// requeueStaleJobsQuery (Crash Recovery):
	// Finds jobs stuck in 'processing' longer than the stale threshold $1
	// (locked_at too old) => the previous worker crashed or was killed mid-processing.
	// Moves the job back to 'queued' with available_at = now() (immediately claimable) and clears locked_at.
	// Does NOT touch attempts: recovery is not a processing attempt.
	requeueStaleJobsQuery = `
	update jobs
	set status = 'queued',
		available_at = now(),
		locked_at = null,
		updated_at = now()
	where status = 'processing'
		and locked_at < now() - $1::interval
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

// ClaimQueuedJob atomically finds a queued job and claims it in a single
// transaction using select ... for update skip locked, so each job is
// only claimed by one worker.
func (r *JobRepository) ClaimQueuedJob(ctx context.Context) (*models.Job, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) // if the transaction is not committed, undo everything

	var job models.Job

	err = tx.QueryRow(ctx, claimQueuedJobQuery).Scan(
		&job.ID,
		&job.Type,
		&job.Payload,
		&job.Status,
		&job.Attempts,
		&job.MaxAttempts,
		&job.AvailableAt,
		&job.LockedAt,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // no job available -> nil, the worker waits for the next round
	}

	if err != nil {
		return nil, err
	}

	err = tx.QueryRow(ctx, claimQueuedJobUpdateQuery, job.ID).Scan(&job.Attempts)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &job, nil
}

func (r *JobRepository) UpdateJobStatus(ctx context.Context, id uuid.UUID, status models.JobStatus, attempts int, availableAt time.Time) error {
	_, err := r.db.Exec(ctx, updateJobStatusQuery, id, status, attempts, availableAt)
	return err
}

// RequeueStaleJobs moves jobs stuck in 'processing' past the stale timeout
// back to 'queued' (with locked_at cleared) so they can be claimed again.
// It does not change attempts.
func (r *JobRepository) RequeueStaleJobs(ctx context.Context, staleAfter time.Duration) (int, error) {
	// A single UPDATE is inherently atomic: running it twice in a row finds no
	// more rows on the second pass (jobs are already 'queued'), so nothing is
	// requeued twice. staleAfter is the stale threshold, e.g. 30s => any job
	// with locked_at < now() - 30s.
	tag, err := r.db.Exec(ctx, requeueStaleJobsQuery, staleAfter)
	if err != nil {
		return 0, err
	}

	return int(tag.RowsAffected()), nil
}
