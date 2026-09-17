package workers

import (
	"context"
	"quocantran/gojob/internal/models"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

// fakeRepository is an in-memory stand-in for repositories.JobRepository.
// The mutex around ClaimQueuedJob/UpdateJobStatus reproduces the atomicity of the
// Postgres "select ... for update skip locked" claim so the concurrency
// invariant (each job claimed by exactly one worker) can be verified without a real database.
type fakeRepository struct {
	mu     sync.Mutex
	order  []uuid.UUID
	jobs   map[uuid.UUID]*models.Job
	claims map[uuid.UUID]int // count how many claims of each job
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		jobs:   make(map[uuid.UUID]*models.Job),
		claims: make(map[uuid.UUID]int),
	}
}

func (f *fakeRepository) addJobs(n int) {
	now := time.Now()
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := 0; i < n; i++ {
		id := uuid.New()
		f.order = append(f.order, id)
		f.jobs[id] = &models.Job{
			ID:          id,
			Type:        "test",
			Payload:     []byte(`{}`),
			Status:      models.JobStatusQueued,
			Attempts:    0,
			MaxAttempts: 3,
			AvailableAt: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	}
}

func (f *fakeRepository) ClaimQueuedJob(ctx context.Context) (*models.Job, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, id := range f.order {
		j := f.jobs[id]
		if j.Status != models.JobStatusQueued || j.AvailableAt.After(time.Now()) {
			continue
		}

		j.Status = models.JobStatusProcessing
		j.Attempts++
		lockedAt := time.Now()
		j.LockedAt = &lockedAt
		f.claims[id]++

		// Return a snapshot, mirroring what Postgres returns to the worker.
		return &models.Job{
			ID:          j.ID,
			Type:        j.Type,
			Payload:     j.Payload,
			Status:      j.Status,
			Attempts:    j.Attempts,
			MaxAttempts: j.MaxAttempts,
			AvailableAt: j.AvailableAt,
			LockedAt:    j.LockedAt,
			CreatedAt:   j.CreatedAt,
			UpdatedAt:   j.UpdatedAt,
		}, nil
	}

	return nil, nil
}

func (f *fakeRepository) UpdateJobStatus(ctx context.Context, id uuid.UUID, status models.JobStatus, attempts int, availableAt time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	j, ok := f.jobs[id]
	if !ok {
		return nil
	}
	j.Status = status
	j.Attempts = attempts
	j.AvailableAt = availableAt
	j.UpdatedAt = time.Now()
	j.LockedAt = nil
	return nil
}

func (f *fakeRepository) completedCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	n := 0
	for _, j := range f.jobs {
		if j.Status == models.JobStatusCompleted {
			n++
		}
	}
	return n
}

func (f *fakeRepository) claimCount(id uuid.UUID) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.claims[id]
}

// fakeProcessor always succeeds so no job is ever retried: therefore a job
// claimed twice would surface as a claim count of > 1.
type fakeProcessor struct{}

func (p *fakeProcessor) Process(ctx context.Context, job *models.Job) error {
	return nil
}

func TestConcurrentClaimsEachJobExactlyOnce(t *testing.T) {
	const (
		jobCount    = 100
		workerCount = 10
	)

	repo := newFakeRepository()
	repo.addJobs(jobCount)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		NewWorkerPool(repo, workerCount, &fakeProcessor{}).Run(ctx)
		close(done)
	}()

	// Wait until all jobs are completed (i.e. claimed and processed once).
	deadline := time.Now().Add(10 * time.Second)
	for repo.completedCount() < jobCount {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %d jobs to complete, completed=%d", jobCount, repo.completedCount())
		}
		time.Sleep(10 * time.Millisecond)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("worker pool did not shut down after cancellation")
	}

	for _, id := range repo.order {
		if got := repo.claimCount(id); got != 1 {
			t.Fatalf("job %s claimed %d times, want exactly 1", id, got)
		}
	}
}
