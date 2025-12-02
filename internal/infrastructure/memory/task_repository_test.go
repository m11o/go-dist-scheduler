package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourname/go-dist-scheduler/internal/domain"
)

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
