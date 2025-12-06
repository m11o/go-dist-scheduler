package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/yourname/go-dist-scheduler/internal/domain"
	"github.com/yourname/go-dist-scheduler/internal/infrastructure/memory"
)

func TestExecutor_RunPendingJob_Success(t *testing.T) {
	ctx := context.Background()
	jobRepo := memory.NewInMemoryJobRepository()
	executor := NewExecutor(jobRepo)

	// Enqueue a pending job
	jobID := uuid.NewString()
	pendingJob := &domain.Job{
		ID:          jobID,
		TaskID:      uuid.NewString(),
		ScheduledAt: time.Now(),
		Status:      domain.JobStatusPending,
	}
	err := jobRepo.Enqueue(ctx, pendingJob)
	assert.NoError(t, err)

	// Run pending jobs
	err = executor.RunPendingJob(ctx)
	assert.NoError(t, err)

	// Verify job status transition
	job, err := jobRepo.FindByID(ctx, jobID)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, domain.JobStatusSuccess, job.Status)
	assert.NotZero(t, job.StartedAt)
	assert.NotZero(t, job.FinishedAt)
}

func TestExecutor_RunPendingJob_FailureOnUpdateToSuccess(t *testing.T) {
	ctx := context.Background()
	jobRepo := memory.NewInMemoryJobRepository()
	executor := NewExecutor(jobRepo)

	// Enqueue a pending job
	jobID := uuid.NewString()
	pendingJob := &domain.Job{
		ID:          jobID,
		TaskID:      uuid.NewString(),
		ScheduledAt: time.Now(),
		Status:      domain.JobStatusPending,
	}
	err := jobRepo.Enqueue(ctx, pendingJob)
	assert.NoError(t, err)

	// Inject an error to be returned on the second UpdateStatus call
	jobRepo.SetUpdateError(domain.JobStatusSuccess, errors.New("failed to update job"))

	// Run pending jobs
	err = executor.RunPendingJob(ctx)
	assert.Error(t, err)

	// Verify job status transition to Failed
	job, err := jobRepo.FindByID(ctx, jobID)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, domain.JobStatusFailed, job.Status)
	assert.NotZero(t, job.StartedAt)
	assert.NotZero(t, job.FinishedAt)
}

func TestExecutor_RunPendingJob_DequeueError(t *testing.T) {
	ctx := context.Background()
	jobRepo := memory.NewInMemoryJobRepository()
	executor := NewExecutor(jobRepo)

	// Inject an error to be returned on Dequeue
	jobRepo.SetDequeueError(errors.New("failed to dequeue job"))

	err := executor.RunPendingJob(ctx)
	assert.Error(t, err)
}

func TestExecutor_RunPendingJob_UpdateStatusToRunningError(t *testing.T) {
	ctx := context.Background()
	jobRepo := memory.NewInMemoryJobRepository()
	executor := NewExecutor(jobRepo)

	// Enqueue a pending job
	jobID := uuid.NewString()
	pendingJob := &domain.Job{
		ID:          jobID,
		TaskID:      uuid.NewString(),
		ScheduledAt: time.Now(),
		Status:      domain.JobStatusPending,
	}
	err := jobRepo.Enqueue(ctx, pendingJob)
	assert.NoError(t, err)

	// Inject an error to be returned on the first UpdateStatus call
	jobRepo.SetUpdateError(domain.JobStatusRunning, errors.New("failed to update job"))

	err = executor.RunPendingJob(ctx)
	assert.Error(t, err)
}

func TestExecutor_RunPendingJob_NoPendingJobs(t *testing.T) {
	ctx := context.Background()
	jobRepo := memory.NewInMemoryJobRepository()
	executor := NewExecutor(jobRepo)

	err := executor.RunPendingJob(ctx)
	assert.NoError(t, err)
}
