package usecase

import (
	"context"
	"log"
	"time"

	"github.com/yourname/go-dist-scheduler/internal/domain"
)

// Executor is responsible for executing pending jobs.
type Executor struct {
	jobRepo domain.JobRepository
}

// NewExecutor creates a new Executor.
func NewExecutor(jobRepo domain.JobRepository) *Executor {
	return &Executor{jobRepo: jobRepo}
}

func (e *Executor) RunPendingJobs(ctx context.Context) error {
	job, err := e.jobRepo.Dequeue(ctx)
	if err != nil {
		return err
	}
	if job == nil {
		return nil
	}

	job.Status = domain.JobStatusRunning
	job.StartedAt = time.Now()
	job.UpdatedAt = time.Now()
	if err := e.jobRepo.UpdateStatus(ctx, job); err != nil {
		return err
	}

	log.Printf("Executing Job ID: %s", job.ID)
	time.Sleep(10 * time.Millisecond) // Simulate work

	job.Status = domain.JobStatusSuccess
	job.FinishedAt = time.Now()
	job.UpdatedAt = time.Now()
	if err := e.jobRepo.UpdateStatus(ctx, job); err != nil {
		job.Status = domain.JobStatusFailed
		job.FinishedAt = time.Now()
		job.UpdatedAt = time.Now()
		if errUpdateFailed := e.jobRepo.UpdateStatus(ctx, job); errUpdateFailed != nil {
			log.Printf("failed to update job %s to Failed status after another error: %v", job.ID, errUpdateFailed)
		}
		return err
	}

	return nil
}
