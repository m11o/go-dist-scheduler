
package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yourname/go-dist-scheduler/internal/domain"
)

func TestInMemoryTaskRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryTaskRepository()

	// Test Save and FindByID
	task1 := &domain.Task{ID: "1", Name: "Test Task 1", Status: domain.TaskStatusActive}
	err := repo.Save(ctx, task1)
	assert.NoError(t, err)

	foundTask, err := repo.FindByID(ctx, "1")
	assert.NoError(t, err)
	assert.Equal(t, task1, foundTask)

	// Test FindAllActive
	task2 := &domain.Task{ID: "2", Name: "Test Task 2", Status: domain.TaskStatusPaused}
	repo.Save(ctx, task2)

	activeTasks, err := repo.FindAllActive(ctx)
	assert.NoError(t, err)
	assert.Len(t, activeTasks, 1)
	assert.Equal(t, task1, activeTasks[0])
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
	assert.Equal(t, job1, dequeuedJob)

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
	assert.Equal(t, dequeuedJob2.FinishedAt, foundJob.FinishedAt)
}
