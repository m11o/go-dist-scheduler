package postgres_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourname/go-dist-scheduler/internal/domain"
	"github.com/yourname/go-dist-scheduler/internal/infrastructure/postgres"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	// Skip test if DB_PASSWORD is not set (not in CI/integration test environment)
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		t.Skip("DB_PASSWORD not set, skipping integration test")
	}

	// Get database configuration from environment variables
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "scheduler"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "scheduler"
	}
	dbSSLMode := os.Getenv("DB_SSLMODE")
	if dbSSLMode == "" {
		dbSSLMode = "disable"
	}

	dsn := "host=" + dbHost + " port=" + dbPort + " user=" + dbUser +
		" password=" + dbPassword + " dbname=" + dbName + " sslmode=" + dbSSLMode

	db, err := postgres.NewClient(dsn)
	require.NoError(t, err, "failed to connect to database")

	// Clean up function to close DB and clean test data
	cleanup := func() {
		// Clean up test data
		if _, err := db.Exec("DELETE FROM tasks"); err != nil {
			t.Logf("warning: failed to clean up tasks: %v", err)
		}
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close database connection: %v", err)
		}
	}

	return db, cleanup
}

func TestTaskRepository_Save_Insert(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := postgres.NewTaskRepository(db)
	ctx := context.Background()

	task := &domain.Task{
		ID:             uuid.NewString(),
		Name:           "Test Task 1",
		CronExpression: "* * * * *",
		Payload: domain.HTTPRequestInfo{
			URL:     "http://example.com",
			Method:  "POST",
			Headers: map[string]string{"Content-Type": "application/json"},
			Body:    []byte(`{"key":"value"}`),
		},
		Status:    domain.TaskStatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	err := repo.Save(ctx, task)
	assert.NoError(t, err)

	// Verify the task was saved
	savedTask, err := repo.FindByID(ctx, task.ID)
	assert.NoError(t, err)
	require.NotNil(t, savedTask)
	assert.Equal(t, task.ID, savedTask.ID)
	assert.Equal(t, task.Name, savedTask.Name)
	assert.Equal(t, task.CronExpression, savedTask.CronExpression)
	assert.Equal(t, task.Status, savedTask.Status)
	assert.Equal(t, task.Payload.URL, savedTask.Payload.URL)
	assert.Equal(t, task.Payload.Method, savedTask.Payload.Method)
	assert.Equal(t, task.Payload.Headers, savedTask.Payload.Headers)
	assert.Equal(t, task.Payload.Body, savedTask.Payload.Body)
}

func TestTaskRepository_Save_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := postgres.NewTaskRepository(db)
	ctx := context.Background()

	// Insert initial task
	task := &domain.Task{
		ID:             uuid.NewString(),
		Name:           "Test Task 2",
		CronExpression: "* * * * *",
		Payload: domain.HTTPRequestInfo{
			URL:    "http://example.com",
			Method: "GET",
		},
		Status:    domain.TaskStatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	err := repo.Save(ctx, task)
	require.NoError(t, err)

	// Update the task
	task.Name = "Updated Task 2"
	task.UpdatedAt = time.Now().UTC()

	err = repo.Save(ctx, task)
	assert.NoError(t, err)

	// Verify the update
	savedTask, err := repo.FindByID(ctx, task.ID)
	assert.NoError(t, err)
	require.NotNil(t, savedTask)
	assert.Equal(t, "Updated Task 2", savedTask.Name)
}

func TestTaskRepository_FindByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := postgres.NewTaskRepository(db)
	ctx := context.Background()

	task, err := repo.FindByID(ctx, uuid.NewString())
	assert.NoError(t, err)
	assert.Nil(t, task)
}

func TestTaskRepository_FindAllActive(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := postgres.NewTaskRepository(db)
	ctx := context.Background()

	activeTaskID := uuid.NewString()

	// Insert active task
	activeTask := &domain.Task{
		ID:             activeTaskID,
		Name:           "Active Task",
		CronExpression: "* * * * *",
		Payload: domain.HTTPRequestInfo{
			URL:    "http://example.com",
			Method: "GET",
		},
		Status:    domain.TaskStatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	err := repo.Save(ctx, activeTask)
	require.NoError(t, err)

	// Insert paused task
	pausedTask := &domain.Task{
		ID:             uuid.NewString(),
		Name:           "Paused Task",
		CronExpression: "* * * * *",
		Payload: domain.HTTPRequestInfo{
			URL:    "http://example.com",
			Method: "GET",
		},
		Status:    domain.TaskStatusPaused,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	err = repo.Save(ctx, pausedTask)
	require.NoError(t, err)

	// Find all active tasks
	tasks, err := repo.FindAllActive(ctx)
	assert.NoError(t, err)
	require.NotNil(t, tasks)

	// Filter to only test task
	var testActiveTasks []*domain.Task
	for _, task := range tasks {
		if task.ID == activeTaskID {
			testActiveTasks = append(testActiveTasks, task)
		}
	}

	assert.Len(t, testActiveTasks, 1)
	assert.Equal(t, activeTaskID, testActiveTasks[0].ID)
	assert.Equal(t, domain.TaskStatusActive, testActiveTasks[0].Status)
}

func TestTaskRepository_SaveAndRetrieve_WithLastCheckedAt(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := postgres.NewTaskRepository(db)
	ctx := context.Background()

	lastChecked := time.Now().UTC().Add(-1 * time.Hour)
	task := &domain.Task{
		ID:             uuid.NewString(),
		Name:           "Test Task with LastCheckedAt",
		CronExpression: "* * * * *",
		Payload: domain.HTTPRequestInfo{
			URL:    "http://example.com",
			Method: "GET",
		},
		Status:        domain.TaskStatusActive,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
		LastCheckedAt: lastChecked,
	}

	err := repo.Save(ctx, task)
	require.NoError(t, err)

	// Retrieve and verify LastCheckedAt
	savedTask, err := repo.FindByID(ctx, task.ID)
	assert.NoError(t, err)
	require.NotNil(t, savedTask)
	assert.False(t, savedTask.LastCheckedAt.IsZero())
	// Allow small time difference due to database precision
	assert.WithinDuration(t, lastChecked, savedTask.LastCheckedAt, time.Second)
}

func TestTaskRepository_PayloadEdgeCases(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := postgres.NewTaskRepository(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		payload domain.HTTPRequestInfo
	}{
		{
			name: "nil body",
			payload: domain.HTTPRequestInfo{
				URL:     "http://example.com",
				Method:  "POST",
				Headers: map[string]string{"Content-Type": "application/json"},
				Body:    nil,
			},
		},
		{
			name: "empty body",
			payload: domain.HTTPRequestInfo{
				URL:     "http://example.com",
				Method:  "POST",
				Headers: map[string]string{"Content-Type": "application/json"},
				Body:    []byte{},
			},
		},
		{
			name: "nil headers",
			payload: domain.HTTPRequestInfo{
				URL:     "http://example.com",
				Method:  "POST",
				Headers: nil,
				Body:    []byte(`{"key":"value"}`),
			},
		},
		{
			name: "empty headers",
			payload: domain.HTTPRequestInfo{
				URL:     "http://example.com",
				Method:  "POST",
				Headers: map[string]string{},
				Body:    []byte(`{"key":"value"}`),
			},
		},
		{
			name: "empty URL",
			payload: domain.HTTPRequestInfo{
				URL:     "",
				Method:  "POST",
				Headers: map[string]string{"Content-Type": "application/json"},
				Body:    []byte(`{"key":"value"}`),
			},
		},
		{
			name: "empty Method",
			payload: domain.HTTPRequestInfo{
				URL:     "http://example.com",
				Method:  "",
				Headers: map[string]string{"Content-Type": "application/json"},
				Body:    []byte(`{"key":"value"}`),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &domain.Task{
				ID:             uuid.NewString(),
				Name:           "Edge Case Task: " + tt.name,
				CronExpression: "* * * * *",
				Payload:        tt.payload,
				Status:         domain.TaskStatusActive,
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
			}

			// Save the task
			err := repo.Save(ctx, task)
			require.NoError(t, err, "failed to save task with %s", tt.name)

			// Retrieve and verify
			savedTask, err := repo.FindByID(ctx, task.ID)
			require.NoError(t, err, "failed to retrieve task with %s", tt.name)
			require.NotNil(t, savedTask)

			assert.Equal(t, tt.payload.URL, savedTask.Payload.URL)
			assert.Equal(t, tt.payload.Method, savedTask.Payload.Method)
			assert.Equal(t, tt.payload.Body, savedTask.Payload.Body)

			// Compare headers (handling nil vs empty map)
			if tt.payload.Headers == nil {
				assert.Nil(t, savedTask.Payload.Headers)
			} else {
				assert.Equal(t, tt.payload.Headers, savedTask.Payload.Headers)
			}
		})
	}
}

func TestTaskRepository_ConcurrentUpdates(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := postgres.NewTaskRepository(db)
	ctx := context.Background()

	// Create initial task
	taskID := uuid.NewString()
	task := &domain.Task{
		ID:             taskID,
		Name:           "Concurrent Test Task",
		CronExpression: "* * * * *",
		Payload: domain.HTTPRequestInfo{
			URL:    "http://example.com",
			Method: "GET",
		},
		Status:    domain.TaskStatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	err := repo.Save(ctx, task)
	require.NoError(t, err)

	// Run concurrent updates
	const numGoroutines = 10
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(iteration int) {
			// Retrieve the task
			task, err := repo.FindByID(ctx, taskID)
			if err != nil {
				done <- err
				return
			}
			if task == nil {
				done <- assert.AnError
				return
			}

			// Update the task
			task.Name = task.Name + " - Updated"
			task.UpdatedAt = time.Now().UTC()

			err = repo.Save(ctx, task)
			done <- err
		}(i)
	}

	// Wait for all goroutines to complete
	var successCount, errorCount int
	for i := 0; i < numGoroutines; i++ {
		err := <-done
		if err == nil {
			successCount++
		} else {
			errorCount++
		}
	}

	// All operations should succeed with pessimistic locking
	assert.Equal(t, numGoroutines, successCount, "all concurrent updates should succeed with pessimistic locking")
	assert.Equal(t, 0, errorCount, "no errors should occur with pessimistic locking")

	// Verify final state
	finalTask, err := repo.FindByID(ctx, taskID)
	require.NoError(t, err)
	require.NotNil(t, finalTask)

	// Due to concurrent updates, we just verify the task exists and was updated
	assert.Contains(t, finalTask.Name, "Concurrent Test Task")
	assert.Contains(t, finalTask.Name, "Updated")
}
