package workers

import (
	"context"
	"log"
	"quocantran/gojob/internal/repositories"
	"quocantran/gojob/pkg"
	"time"
)

const (
	// recoveryInterval: how often the RecoveryWorker runs (every 10s).
	recoveryInterval = 10 * time.Second
	// jobTimeout: the stale threshold. A job in 'processing' whose locked_at
	// is older than 30s is considered stuck (worker crashed / was killed /
	// server restarted mid-processing) and gets requeued. The next claim
	// overwrites locked_at, so fresh jobs are never touched by recovery.
	jobTimeout = 30 * time.Second
)

// RecoveryWorker periodically requeues jobs stuck in 'processing' past the
// stale timeout (e.g. a worker crashed, was killed, or the server restarted
// mid-processing), so another worker can claim and reprocess them. Recovery
// never increments attempts.
type RecoveryWorker struct {
	repository *repositories.JobRepository
}

func NewRecoveryWorker(repository *repositories.JobRepository) *RecoveryWorker {
	return &RecoveryWorker{repository: repository}
}

// Run starts the RecoveryWorker on its own goroutine as a ticker-driven loop.
func (r *RecoveryWorker) Run(ctx context.Context) {
	log.Printf("RecoveryWorker is starting ... (interval=%v, stale timeout=%v)\n", recoveryInterval, jobTimeout)

	ticker := time.NewTicker(recoveryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("RecoveryWorker is shutting down ...")
			return
		case <-ticker.C:
			r.recoverStaleJobs(ctx)
		}
	}
}

func (r *RecoveryWorker) recoverStaleJobs(ctx context.Context) {
	log.Printf("RecoveryWorker: checking for stale jobs at %s ...\n", pkg.FormatVN(time.Now()))

	// RequeueStaleJobs performs "find stale + requeue" in a single UPDATE and
	// returns how many jobs were recovered. attempts stay unchanged.
	count, err := r.repository.RequeueStaleJobs(ctx, jobTimeout)
	if err != nil {
		log.Printf("RecoveryWorker: failed to requeue stale jobs: %v\n", err)
		return
	}

	if count > 0 {
		log.Printf("RecoveryWorker: requeued %d stale job(s) back to 'queued' (was stuck in 'processing' > %v)\n", count, jobTimeout)
	} else {
		log.Printf("RecoveryWorker: no stale job found (all processing jobs are younger than %v)\n", jobTimeout)
	}
}
