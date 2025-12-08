package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourname/go-dist-scheduler/internal/domain"
)

func TestInMemoryTaskRepository_Copy(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryTaskRepository()

	task := &domain.Task{
		ID:   "1",
		Name: "Original Task",
		Payload: domain.HTTPRequestInfo{
			Headers: map[string]string{"Content-Type": "application/json"},
			Body:    []byte(`{"key":"value"}`),
		},
		Status: domain.TaskStatusActive,
	}

	_ = repo.Save(ctx, task)

	// Modify original task after saving
	task.Name = "Modified Task"
	task.Payload.Headers["X-Test"] = "true"
	task.Payload.Body[8] = 'X'

	foundTask, _ := repo.FindByID(ctx, "1")

	// Check that the found task is a copy and not the modified original
	assert.Equal(t, "Original Task", foundTask.Name)
	assert.Equal(t, "application/json", foundTask.Payload.Headers["Content-Type"])
	assert.NotContains(t, foundTask.Payload.Headers, "X-Test")
	assert.Equal(t, `{"key":"value"}`, string(foundTask.Payload.Body))

	// Modify the found task
	foundTask.Name = "Modified Found Task"
	internalTask, _ := repo.FindByID(ctx, "1") // Re-fetch to check internal state

	// Check that modifying the returned task doesn't affect the stored one
	assert.Equal(t, "Original Task", internalTask.Name)
}

func TestInMemoryJobRepository_Copy(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryJobRepository()

	job := &domain.Job{ID: "1", TaskID: "task1", Status: domain.JobStatusPending}

	_ = repo.Enqueue(ctx, job)
	job.Status = domain.JobStatusRunning // Modify original after enqueue

	dequeuedJob, _ := repo.Dequeue(ctx)
	assert.Equal(t, domain.JobStatusPending, dequeuedJob.Status)
}

func TestInMemoryJobRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryJobRepository()

	// Test Enqueue and Dequeue
	job1 := &domain.Job{ID: "1", TaskID: "task1"}
	err := repo.Enqueue(ctx, job1)
	assert.NoError(t, err)

	dequeuedJob, err := repo.Dequeue(ctx)
	assert.NoError(t, err)
	assert.EqualValues(t, job1, dequeuedJob)

	// Test Dequeue from empty queue
	dequeuedJob, err = repo.Dequeue(ctx)
	assert.NoError(t, err)
	assert.Nil(t, dequeuedJob)
}

func TestInMemoryJobRepository_UpdateStatus(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryJobRepository()

	// Test UpdateStatus for Running status
	job1 := &domain.Job{ID: "job1", TaskID: "task1", Status: domain.JobStatusPending}
	err := repo.Enqueue(ctx, job1)
	assert.NoError(t, err)

	err = repo.UpdateStatus(ctx, "job1", domain.JobStatusRunning)
	assert.NoError(t, err)

	// Dequeue and verify status was updated
	dequeuedJob, err := repo.Dequeue(ctx)
	assert.NoError(t, err)
	assert.Equal(t, domain.JobStatusRunning, dequeuedJob.Status)
	assert.False(t, dequeuedJob.StartedAt.IsZero(), "StartedAt should be set when marking as Running")

	// Test UpdateStatus for Success status
	job2 := &domain.Job{ID: "job2", TaskID: "task2", Status: domain.JobStatusPending}
	err = repo.Enqueue(ctx, job2)
	assert.NoError(t, err)

	err = repo.UpdateStatus(ctx, "job2", domain.JobStatusSuccess)
	assert.NoError(t, err)

	dequeuedJob2, err := repo.Dequeue(ctx)
	assert.NoError(t, err)
	assert.Equal(t, domain.JobStatusSuccess, dequeuedJob2.Status)
	assert.False(t, dequeuedJob2.FinishedAt.IsZero(), "FinishedAt should be set when marking as Success")

	// Test UpdateStatus for Failed status
	job3 := &domain.Job{ID: "job3", TaskID: "task3", Status: domain.JobStatusPending}
	err = repo.Enqueue(ctx, job3)
	assert.NoError(t, err)

	err = repo.UpdateStatus(ctx, "job3", domain.JobStatusFailed)
	assert.NoError(t, err)

	dequeuedJob3, err := repo.Dequeue(ctx)
	assert.NoError(t, err)
	assert.Equal(t, domain.JobStatusFailed, dequeuedJob3.Status)
	assert.False(t, dequeuedJob3.FinishedAt.IsZero(), "FinishedAt should be set when marking as Failed")

	// Test UpdateStatus on non-existent job (should not error)
	err = repo.UpdateStatus(ctx, "nonexistent", domain.JobStatusSuccess)
	assert.NoError(t, err)
}

func TestInMemoryTaskRepository_Save_Conflict(t *testing.T) {
	repo := NewInMemoryTaskRepository()
	ctx := context.Background()

	// Initial save
	task1 := &domain.Task{ID: "task1", Version: 1}
	err := repo.Save(ctx, task1)
	assert.NoError(t, err)

	// Try to save with the same version again (should fail)
	task2 := &domain.Task{ID: "task1", Version: 1}
	err = repo.Save(ctx, task2)
	assert.ErrorIs(t, err, domain.ErrConflict)

	// Save with the correct next version (should succeed)
	task3 := &domain.Task{ID: "task1", Version: 2}
	err = repo.Save(ctx, task3)
	assert.NoError(t, err)

	// Check final version
	savedTask, err := repo.FindByID(ctx, "task1")
	assert.NoError(t, err)
	assert.Equal(t, 2, savedTask.Version)
}
