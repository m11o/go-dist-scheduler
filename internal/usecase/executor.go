package usecase

import (
	"context"
	"log"
	"time"

	"github.com/yourname/go-dist-scheduler/internal/domain"
)

// Executor is responsible for executing pending jobs.
type Executor struct {
	jobRepo  domain.JobRepository
	jobQueue domain.JobQueue
}

// NewExecutor creates a new Executor.
func NewExecutor(jobRepo domain.JobRepository, jobQueue domain.JobQueue) *Executor {
	return &Executor{
		jobRepo:  jobRepo,
		jobQueue: jobQueue,
	}
}

func (e *Executor) RunPendingJob(ctx context.Context) error {
	job, err := e.jobQueue.Dequeue(ctx)
	if err != nil {
		return err
	}
	if job == nil {
		return nil
	}

	job.MarkAsRunning()
	if err := e.jobRepo.Update(ctx, job); err != nil {
		return err
	}

	log.Printf("Executing Job ID: %s", job.ID)
	time.Sleep(10 * time.Millisecond) // Simulate work

	job.MarkAsSuccess()
	if err := e.jobRepo.Update(ctx, job); err != nil {
		log.Printf("failed to update job %s to Success status: %v", job.ID, err)
		return err
	}

	return nil
}
