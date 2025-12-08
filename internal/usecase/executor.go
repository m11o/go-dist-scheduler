package usecase

import (
	"context"
	"log"
	"time"

	"github.com/yourname/go-dist-scheduler/internal/domain"
)

// Executor は、ペンディング中のジョブを実行する責務を担当します。
type Executor struct {
	jobRepo domain.JobRepository
}

// NewExecutor は新しいExecutorインスタンスを生成します。
func NewExecutor(jobRepo domain.JobRepository) *Executor {
	return &Executor{
		jobRepo: jobRepo,
	}
}

// RunPendingJob は、キューから1つのジョブをデキューして実行します。
func (e *Executor) RunPendingJob(ctx context.Context) error {
	job, err := e.jobRepo.Dequeue(ctx)
	if err != nil {
		return err
	}
	if job == nil {
		return nil
	}

	// Update status to Running
	if err := e.jobRepo.UpdateStatus(ctx, job.ID, domain.JobStatusRunning); err != nil {
		return err
	}

	log.Printf("Executing Job ID: %s", job.ID)
	time.Sleep(10 * time.Millisecond) // Simulate work

	// Update status to Success
	if err := e.jobRepo.UpdateStatus(ctx, job.ID, domain.JobStatusSuccess); err != nil {
		log.Printf("failed to update job %s to Success status: %v", job.ID, err)
		return err
	}

	return nil
}
