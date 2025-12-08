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
}

type dequeueErrorJobRepository struct {
	memory.InMemoryJobRepository
}

func (r *dequeueErrorJobRepository) Dequeue(ctx context.Context) (*domain.Job, error) {
	return nil, errors.New("failed to dequeue job")
}

func newDequeueErrorJobRepository() *dequeueErrorJobRepository {
	return &dequeueErrorJobRepository{
		InMemoryJobRepository: *memory.NewInMemoryJobRepository(),
	}
}

func TestExecutor_RunPendingJob_DequeueError(t *testing.T) {
	ctx := context.Background()
	jobRepo := newDequeueErrorJobRepository()
	executor := NewExecutor(jobRepo)

	err := executor.RunPendingJob(ctx)
	assert.Error(t, err)
}

func TestExecutor_RunPendingJob_NoPendingJobs(t *testing.T) {
	ctx := context.Background()
	jobRepo := memory.NewInMemoryJobRepository()
	executor := NewExecutor(jobRepo)

	err := executor.RunPendingJob(ctx)
	assert.NoError(t, err)
}

type updateStatusErrorJobRepository struct {
	memory.InMemoryJobRepository
	updateCount int
}

func (r *updateStatusErrorJobRepository) UpdateStatus(ctx context.Context, jobID string, status domain.JobStatus) error {
	r.updateCount++
	if r.updateCount == 1 {
		// First call (Running status) fails
		return errors.New("failed to update status to Running")
	}
	return r.InMemoryJobRepository.UpdateStatus(ctx, jobID, status)
}

func newUpdateStatusErrorJobRepository() *updateStatusErrorJobRepository {
	return &updateStatusErrorJobRepository{
		InMemoryJobRepository: *memory.NewInMemoryJobRepository(),
	}
}

func TestExecutor_RunPendingJob_UpdateStatusToRunningError(t *testing.T) {
	ctx := context.Background()
	jobRepo := newUpdateStatusErrorJobRepository()
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

	err = executor.RunPendingJob(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update status to Running")
}

type updateStatusToSuccessErrorJobRepository struct {
	memory.InMemoryJobRepository
	updateCount int
}

func (r *updateStatusToSuccessErrorJobRepository) UpdateStatus(ctx context.Context, jobID string, status domain.JobStatus) error {
	r.updateCount++
	if r.updateCount == 2 && status == domain.JobStatusSuccess {
		// Second call (Success status) fails
		return errors.New("failed to update status to Success")
	}
	return r.InMemoryJobRepository.UpdateStatus(ctx, jobID, status)
}

func newUpdateStatusToSuccessErrorJobRepository() *updateStatusToSuccessErrorJobRepository {
	return &updateStatusToSuccessErrorJobRepository{
		InMemoryJobRepository: *memory.NewInMemoryJobRepository(),
	}
}

func TestExecutor_RunPendingJob_UpdateStatusToSuccessError(t *testing.T) {
	ctx := context.Background()
	jobRepo := newUpdateStatusToSuccessErrorJobRepository()
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

	err = executor.RunPendingJob(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update status to Success")
}
