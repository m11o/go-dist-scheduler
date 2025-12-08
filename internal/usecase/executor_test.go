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
	jobQueue := memory.NewInMemoryJobQueue()
	executor := NewExecutor(jobRepo, jobQueue)

	// Enqueue a pending job
	jobID := uuid.NewString()
	pendingJob := &domain.Job{
		ID:          jobID,
		TaskID:      uuid.NewString(),
		ScheduledAt: time.Now(),
		Status:      domain.JobStatusPending,
	}
	err := jobRepo.Save(ctx, pendingJob)
	assert.NoError(t, err)
	err = jobQueue.Enqueue(ctx, pendingJob)
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

type updateErrorJobRepository struct {
	memory.InMemoryJobRepository
}

func (r *updateErrorJobRepository) Update(ctx context.Context, job *domain.Job) error {
	return errors.New("failed to update job")
}

func newUpdateErrorJobRepository() *updateErrorJobRepository {
	return &updateErrorJobRepository{
		InMemoryJobRepository: *memory.NewInMemoryJobRepository(),
	}
}

func TestExecutor_RunPendingJob_UpdateStatusToRunningError(t *testing.T) {
	ctx := context.Background()
	jobRepo := newUpdateErrorJobRepository()
	jobQueue := memory.NewInMemoryJobQueue()
	executor := NewExecutor(jobRepo, jobQueue)

	// Enqueue a pending job
	jobID := uuid.NewString()
	pendingJob := &domain.Job{
		ID:          jobID,
		TaskID:      uuid.NewString(),
		ScheduledAt: time.Now(),
		Status:      domain.JobStatusPending,
	}
	err := jobQueue.Enqueue(ctx, pendingJob)
	assert.NoError(t, err)

	err = executor.RunPendingJob(ctx)
	assert.Error(t, err)
}

type updateToSuccessErrorJobRepository struct {
	memory.InMemoryJobRepository
}

func (r *updateToSuccessErrorJobRepository) Update(ctx context.Context, job *domain.Job) error {
	if job.Status == domain.JobStatusSuccess {
		return errors.New("failed to update job")
	}
	return r.InMemoryJobRepository.Update(ctx, job)
}

func newUpdateToSuccessErrorJobRepository() *updateToSuccessErrorJobRepository {
	return &updateToSuccessErrorJobRepository{
		InMemoryJobRepository: *memory.NewInMemoryJobRepository(),
	}
}

func TestExecutor_RunPendingJob_FailureOnUpdateToSuccess(t *testing.T) {
	ctx := context.Background()
	jobRepo := newUpdateToSuccessErrorJobRepository()
	jobQueue := memory.NewInMemoryJobQueue()
	executor := NewExecutor(jobRepo, jobQueue)

	// Enqueue a pending job
	jobID := uuid.NewString()
	pendingJob := &domain.Job{
		ID:          jobID,
		TaskID:      uuid.NewString(),
		ScheduledAt: time.Now(),
		Status:      domain.JobStatusPending,
	}
	err := jobRepo.Save(ctx, pendingJob)
	assert.NoError(t, err)
	err = jobQueue.Enqueue(ctx, pendingJob)
	assert.NoError(t, err)

	// Run pending jobs
	err = executor.RunPendingJob(ctx)
	assert.Error(t, err)

	// Verify job status remains Running
	job, err := jobRepo.FindByID(ctx, jobID)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, domain.JobStatusRunning, job.Status)
}

type dequeueErrorJobQueue struct {
	memory.InMemoryJobQueue
}

func (q *dequeueErrorJobQueue) Dequeue(ctx context.Context) (*domain.Job, error) {
	return nil, errors.New("failed to dequeue job")
}

func newDequeueErrorJobQueue() *dequeueErrorJobQueue {
	return &dequeueErrorJobQueue{
		InMemoryJobQueue: *memory.NewInMemoryJobQueue(),
	}
}

func TestExecutor_RunPendingJob_DequeueError(t *testing.T) {
	ctx := context.Background()
	jobRepo := memory.NewInMemoryJobRepository()
	jobQueue := newDequeueErrorJobQueue()
	executor := NewExecutor(jobRepo, jobQueue)

	err := executor.RunPendingJob(ctx)
	assert.Error(t, err)
}

func TestExecutor_RunPendingJob_NoPendingJobs(t *testing.T) {
	ctx := context.Background()
	jobRepo := memory.NewInMemoryJobRepository()
	jobQueue := memory.NewInMemoryJobQueue()
	executor := NewExecutor(jobRepo, jobQueue)

	err := executor.RunPendingJob(ctx)
	assert.NoError(t, err)
}
