package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourname/go-dist-scheduler/internal/domain"
	"github.com/yourname/go-dist-scheduler/internal/infrastructure/postgres"
)

func TestJobRepository_Enqueue(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := postgres.NewJobRepository(db)
	ctx := context.Background()

	job := &domain.Job{
		ID:          uuid.NewString(),
		TaskID:      uuid.NewString(),
		ScheduledAt: time.Now().UTC(),
		Status:      domain.JobStatusPending,
		RetryCount:  0,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	err := repo.Enqueue(ctx, job)
	assert.NoError(t, err)

	// Verify the job was enqueued by attempting to dequeue it
	dequeuedJob, err := repo.Dequeue(ctx)
	assert.NoError(t, err)
	require.NotNil(t, dequeuedJob)
	assert.Equal(t, job.ID, dequeuedJob.ID)
	assert.Equal(t, job.TaskID, dequeuedJob.TaskID)
	assert.Equal(t, job.Status, dequeuedJob.Status)

	// Clean up
	_, err = db.ExecContext(ctx, "DELETE FROM jobs WHERE id = $1", job.ID)
	require.NoError(t, err)
}

func TestJobRepository_Dequeue_EmptyQueue(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := postgres.NewJobRepository(db)
	ctx := context.Background()

	// Clean up any existing jobs
	_, err := db.ExecContext(ctx, "DELETE FROM jobs")
	require.NoError(t, err)

	job, err := repo.Dequeue(ctx)
	assert.NoError(t, err)
	assert.Nil(t, job)
}

func TestJobRepository_Dequeue_OrderByScheduledAt(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := postgres.NewJobRepository(db)
	ctx := context.Background()

	// Clean up any existing jobs
	_, err := db.ExecContext(ctx, "DELETE FROM jobs")
	require.NoError(t, err)

	now := time.Now().UTC()

	// Create jobs with different scheduled times
	job1 := &domain.Job{
		ID:          uuid.NewString(),
		TaskID:      uuid.NewString(),
		ScheduledAt: now.Add(2 * time.Hour),
		Status:      domain.JobStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	job2 := &domain.Job{
		ID:          uuid.NewString(),
		TaskID:      uuid.NewString(),
		ScheduledAt: now.Add(1 * time.Hour),
		Status:      domain.JobStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	job3 := &domain.Job{
		ID:          uuid.NewString(),
		TaskID:      uuid.NewString(),
		ScheduledAt: now.Add(3 * time.Hour),
		Status:      domain.JobStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Enqueue jobs in random order
	err = repo.Enqueue(ctx, job1)
	require.NoError(t, err)
	err = repo.Enqueue(ctx, job3)
	require.NoError(t, err)
	err = repo.Enqueue(ctx, job2)
	require.NoError(t, err)

	// Dequeue should return job2 first (earliest scheduled_at)
	dequeuedJob, err := repo.Dequeue(ctx)
	assert.NoError(t, err)
	require.NotNil(t, dequeuedJob)
	assert.Equal(t, job2.ID, dequeuedJob.ID)

	// Clean up
	_, err = db.ExecContext(ctx, "DELETE FROM jobs WHERE id IN ($1, $2, $3)", job1.ID, job2.ID, job3.ID)
	require.NoError(t, err)
}

func TestJobRepository_UpdateStatus_ToRunning(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := postgres.NewJobRepository(db)
	ctx := context.Background()

	job := &domain.Job{
		ID:          uuid.NewString(),
		TaskID:      uuid.NewString(),
		ScheduledAt: time.Now().UTC(),
		Status:      domain.JobStatusPending,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	err := repo.Enqueue(ctx, job)
	require.NoError(t, err)

	// Update status to Running
	err = repo.UpdateStatus(ctx, job.ID, domain.JobStatusRunning)
	assert.NoError(t, err)

	// Verify the status was updated
	var status int
	var startedAt, finishedAt *time.Time
	err = db.QueryRowContext(ctx, "SELECT status, started_at, finished_at FROM jobs WHERE id = $1", job.ID).Scan(&status, &startedAt, &finishedAt)
	require.NoError(t, err)
	assert.Equal(t, int(domain.JobStatusRunning), status)
	assert.NotNil(t, startedAt)
	assert.Nil(t, finishedAt)

	// Clean up
	_, err = db.ExecContext(ctx, "DELETE FROM jobs WHERE id = $1", job.ID)
	require.NoError(t, err)
}

func TestJobRepository_UpdateStatus_ToSuccess(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := postgres.NewJobRepository(db)
	ctx := context.Background()

	job := &domain.Job{
		ID:          uuid.NewString(),
		TaskID:      uuid.NewString(),
		ScheduledAt: time.Now().UTC(),
		Status:      domain.JobStatusPending,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	err := repo.Enqueue(ctx, job)
	require.NoError(t, err)

	// Update status to Running first
	err = repo.UpdateStatus(ctx, job.ID, domain.JobStatusRunning)
	require.NoError(t, err)

	// Update status to Success
	err = repo.UpdateStatus(ctx, job.ID, domain.JobStatusSuccess)
	assert.NoError(t, err)

	// Verify the status was updated
	var status int
	var startedAt, finishedAt *time.Time
	err = db.QueryRowContext(ctx, "SELECT status, started_at, finished_at FROM jobs WHERE id = $1", job.ID).Scan(&status, &startedAt, &finishedAt)
	require.NoError(t, err)
	assert.Equal(t, int(domain.JobStatusSuccess), status)
	assert.NotNil(t, startedAt)
	assert.NotNil(t, finishedAt)

	// Clean up
	_, err = db.ExecContext(ctx, "DELETE FROM jobs WHERE id = $1", job.ID)
	require.NoError(t, err)
}

func TestJobRepository_UpdateStatus_ToFailed(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := postgres.NewJobRepository(db)
	ctx := context.Background()

	job := &domain.Job{
		ID:          uuid.NewString(),
		TaskID:      uuid.NewString(),
		ScheduledAt: time.Now().UTC(),
		Status:      domain.JobStatusPending,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	err := repo.Enqueue(ctx, job)
	require.NoError(t, err)

	// Update status to Running first
	err = repo.UpdateStatus(ctx, job.ID, domain.JobStatusRunning)
	require.NoError(t, err)

	// Update status to Failed
	err = repo.UpdateStatus(ctx, job.ID, domain.JobStatusFailed)
	assert.NoError(t, err)

	// Verify the status was updated
	var status int
	var startedAt, finishedAt *time.Time
	err = db.QueryRowContext(ctx, "SELECT status, started_at, finished_at FROM jobs WHERE id = $1", job.ID).Scan(&status, &startedAt, &finishedAt)
	require.NoError(t, err)
	assert.Equal(t, int(domain.JobStatusFailed), status)
	assert.NotNil(t, startedAt)
	assert.NotNil(t, finishedAt)

	// Clean up
	_, err = db.ExecContext(ctx, "DELETE FROM jobs WHERE id = $1", job.ID)
	require.NoError(t, err)
}

func TestJobRepository_UpdateStatus_JobNotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := postgres.NewJobRepository(db)
	ctx := context.Background()

	// Try to update a non-existent job
	err := repo.UpdateStatus(ctx, uuid.NewString(), domain.JobStatusRunning)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "job not found")
}

func TestJobRepository_ConcurrentDequeue(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := postgres.NewJobRepository(db)
	ctx := context.Background()

	// Clean up any existing jobs
	_, err := db.ExecContext(ctx, "DELETE FROM jobs")
	require.NoError(t, err)

	now := time.Now().UTC()

	// Create multiple pending jobs
	numJobs := 10
	jobIDs := make([]string, numJobs)
	for i := 0; i < numJobs; i++ {
		job := &domain.Job{
			ID:          uuid.NewString(),
			TaskID:      uuid.NewString(),
			ScheduledAt: now,
			Status:      domain.JobStatusPending,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		jobIDs[i] = job.ID
		err := repo.Enqueue(ctx, job)
		require.NoError(t, err)
	}

	// Run concurrent dequeue operations
	const numWorkers = 5
	results := make(chan *domain.Job, numWorkers)
	errors := make(chan error, numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			job, err := repo.Dequeue(ctx)
			if err != nil {
				errors <- err
				return
			}
			results <- job
		}()
	}

	// Collect results
	var dequeuedJobs []*domain.Job
	var dequeuedErrors []error
	for i := 0; i < numWorkers; i++ {
		select {
		case job := <-results:
			if job != nil {
				dequeuedJobs = append(dequeuedJobs, job)
			}
		case err := <-errors:
			dequeuedErrors = append(dequeuedErrors, err)
		}
	}

	// Verify no errors occurred
	assert.Empty(t, dequeuedErrors)

	// Verify we got exactly numWorkers jobs (or less if there aren't enough)
	assert.LessOrEqual(t, len(dequeuedJobs), numWorkers)

	// Verify no duplicate jobs were dequeued
	seenIDs := make(map[string]bool)
	for _, job := range dequeuedJobs {
		assert.False(t, seenIDs[job.ID], "duplicate job ID dequeued: %s", job.ID)
		seenIDs[job.ID] = true
	}

	// Clean up - Delete all test jobs in a single batch operation
	for _, id := range jobIDs {
		if _, err := db.ExecContext(ctx, "DELETE FROM jobs WHERE id = $1", id); err != nil {
			t.Logf("warning: failed to clean up job %s: %v", id, err)
		}
	}
}

func TestJobRepository_StatusUpdateReflectedInDB(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := postgres.NewJobRepository(db)
	ctx := context.Background()

	job := &domain.Job{
		ID:          uuid.NewString(),
		TaskID:      uuid.NewString(),
		ScheduledAt: time.Now().UTC(),
		Status:      domain.JobStatusPending,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	// Enqueue the job
	err := repo.Enqueue(ctx, job)
	require.NoError(t, err)

	// Verify initial status
	var status int
	err = db.QueryRowContext(ctx, "SELECT status FROM jobs WHERE id = $1", job.ID).Scan(&status)
	require.NoError(t, err)
	assert.Equal(t, int(domain.JobStatusPending), status)

	// Update to Running
	err = repo.UpdateStatus(ctx, job.ID, domain.JobStatusRunning)
	require.NoError(t, err)

	// Verify status updated to Running
	err = db.QueryRowContext(ctx, "SELECT status FROM jobs WHERE id = $1", job.ID).Scan(&status)
	require.NoError(t, err)
	assert.Equal(t, int(domain.JobStatusRunning), status)

	// Update to Success
	err = repo.UpdateStatus(ctx, job.ID, domain.JobStatusSuccess)
	require.NoError(t, err)

	// Verify status updated to Success
	err = db.QueryRowContext(ctx, "SELECT status FROM jobs WHERE id = $1", job.ID).Scan(&status)
	require.NoError(t, err)
	assert.Equal(t, int(domain.JobStatusSuccess), status)

	// Clean up
	_, err = db.ExecContext(ctx, "DELETE FROM jobs WHERE id = $1", job.ID)
	require.NoError(t, err)
}
