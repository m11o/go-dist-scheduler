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
