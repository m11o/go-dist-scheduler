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

func (e *Executor) RunPendingJob(ctx context.Context) error {
	job, err := e.jobRepo.Dequeue(ctx)
	if err != nil {
		return err
	}
	if job == nil {
		return nil
	}

	job.MarkAsRunning()

	log.Printf("Executing Job ID: %s", job.ID)
	time.Sleep(10 * time.Millisecond) // Simulate work

	job.MarkAsSuccess()

	return nil
}
