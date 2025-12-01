
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

	repo.Save(ctx, task)

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

	repo.Enqueue(ctx, job)
	job.Status = domain.JobStatusRunning // Modify original after enqueue

	dequeuedJob, _ := repo.Dequeue(ctx)
	assert.Equal(t, domain.JobStatusPending, dequeuedJob.Status)

	dequeuedJob.Status = domain.JobStatusSuccess // Modify dequeued job
	foundJob, _ := repo.FindByID(ctx, "1")
	assert.Equal(t, domain.JobStatusPending, foundJob.Status)
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
	repo.Save(ctx, task2)

	activeTasks, err := repo.FindAllActive(ctx)
	assert.NoError(t, err)
	assert.Len(t, activeTasks, 1)
	assert.EqualValues(t, task1, activeTasks[0]) // Use EqualValues
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
	assert.EqualValues(t, job1, dequeuedJob) // Use EqualValues

	// Test Dequeue from empty queue
	dequeuedJob, err = repo.Dequeue(ctx)
	assert.NoError(t, err)
	assert.Nil(t, dequeuedJob)

	// Test UpdateStatus
	job2 := &domain.Job{ID: "2", TaskID: "task2", Status: domain.JobStatusPending}
	repo.Enqueue(ctx, job2)

	dequeuedJob2, err := repo.Dequeue(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, dequeuedJob2)

	dequeuedJob2.Status = domain.JobStatusSuccess
	dequeuedJob2.FinishedAt = time.Now()
	err = repo.UpdateStatus(ctx, dequeuedJob2)
	assert.NoError(t, err)

	foundJob, err := repo.FindByID(ctx, "2")
	assert.NoError(t, err)
	assert.Equal(t, dequeuedJob2.Status, foundJob.Status)
	assert.WithinDuration(t, dequeuedJob2.FinishedAt, foundJob.FinishedAt, time.Millisecond) // Compare time
}
