package usecase

import (
	"context"
	"log"
	"time"

	"github.com/yourname/go-dist-scheduler/internal/domain"
)

// Executor は、ペンディング中のジョブを実行する責務を担当します。
type Executor struct {
	jobRepo  domain.JobRepository
	jobQueue domain.JobQueue
}

// NewExecutor は新しいExecutorインスタンスを生成します。
func NewExecutor(jobRepo domain.JobRepository, jobQueue domain.JobQueue) *Executor {
	return &Executor{
		jobRepo:  jobRepo,
		jobQueue: jobQueue,
	}
}

func (e *Executor) RunPendingJob(ctx context.Context) error {
	jobID, err := e.jobQueue.Dequeue(ctx)
	if err != nil {
		return err
	}
	if jobID == "" {
		return nil
	}

	// キューから取得したIDを使ってDBからジョブを取得
	job, err := e.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return err
	}
	if job == nil {
		log.Printf("job not found in repository: %s", jobID)
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
