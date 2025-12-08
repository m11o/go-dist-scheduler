
package memory

import (
	"context"
	"testing"
	"time"

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

	_ = repo.Save(ctx, job)
	job.Status = domain.JobStatusRunning // Modify original after save

	foundJob, _ := repo.FindByID(ctx, "1")
	assert.Equal(t, domain.JobStatusPending, foundJob.Status)

	foundJob.Status = domain.JobStatusSuccess // Modify found job
	refetchedJob, _ := repo.FindByID(ctx, "1")
	assert.Equal(t, domain.JobStatusPending, refetchedJob.Status)
}

func TestInMemoryJobQueue_Copy(t *testing.T) {
	ctx := context.Background()
	queue := NewInMemoryJobQueue()

	job := &domain.Job{ID: "1", TaskID: "task1", Status: domain.JobStatusPending}

	_ = queue.Enqueue(ctx, job)
	job.Status = domain.JobStatusRunning // Modify original after enqueue

	dequeuedJob, _ := queue.Dequeue(ctx)
	assert.Equal(t, domain.JobStatusPending, dequeuedJob.Status)
}

func TestInMemoryTaskRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryTaskRepository()

	// Test Save and FindByID
	task1 := &domain.Task{ID: "1", Name: "Test Task 1", Status: domain.TaskStatusActive}
	err := repo.Save(ctx, task1)
	assert.NoError(t, err)

	foundTask, err := repo.FindByID(ctx, "1")
	assert.NoError(t, err)
	assert.EqualValues(t, task1, foundTask) // Use EqualValues for deep comparison

	// Test FindAllActive
	task2 := &domain.Task{ID: "2", Name: "Test Task 2", Status: domain.TaskStatusPaused}
	_ = repo.Save(ctx, task2)

	activeTasks, err := repo.FindAllActive(ctx)
	assert.NoError(t, err)
	assert.Len(t, activeTasks, 1)
	assert.EqualValues(t, task1, activeTasks[0]) // Use EqualValues
}

func TestInMemoryJobRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryJobRepository()

	// Test Save and FindByID
	job1 := &domain.Job{ID: "1", TaskID: "task1", Status: domain.JobStatusPending}
	err := repo.Save(ctx, job1)
	assert.NoError(t, err)

	foundJob, err := repo.FindByID(ctx, "1")
	assert.NoError(t, err)
	assert.EqualValues(t, job1, foundJob)

	// Test Update
	job1.Status = domain.JobStatusSuccess
	job1.FinishedAt = time.Now()
	err = repo.Update(ctx, job1)
	assert.NoError(t, err)

	updatedJob, err := repo.FindByID(ctx, "1")
	assert.NoError(t, err)
	assert.Equal(t, job1.Status, updatedJob.Status)
	assert.WithinDuration(t, job1.FinishedAt, updatedJob.FinishedAt, time.Millisecond)
}

func TestInMemoryJobQueue(t *testing.T) {
	ctx := context.Background()
	queue := NewInMemoryJobQueue()

	// Test Enqueue and Dequeue
	job1 := &domain.Job{ID: "1", TaskID: "task1"}
	err := queue.Enqueue(ctx, job1)
	assert.NoError(t, err)

	dequeuedJob, err := queue.Dequeue(ctx)
	assert.NoError(t, err)
	assert.EqualValues(t, job1, dequeuedJob)

	// Test Dequeue from empty queue
	dequeuedJob, err = queue.Dequeue(ctx)
	assert.NoError(t, err)
	assert.Nil(t, dequeuedJob)
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
