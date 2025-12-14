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
		_, _ = db.Exec("DELETE FROM tasks")
		db.Close()
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
		Version:   1,
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
	assert.Equal(t, task.Version, savedTask.Version)
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
		Version:   1,
	}

	err := repo.Save(ctx, task)
	require.NoError(t, err)

	// Update the task
	task.Name = "Updated Task 2"
	task.Version = 2
	task.UpdatedAt = time.Now().UTC()

	err = repo.Save(ctx, task)
	assert.NoError(t, err)

	// Verify the update
	savedTask, err := repo.FindByID(ctx, task.ID)
	assert.NoError(t, err)
	require.NotNil(t, savedTask)
	assert.Equal(t, "Updated Task 2", savedTask.Name)
	assert.Equal(t, 2, savedTask.Version)
}

func TestTaskRepository_Save_Conflict(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := postgres.NewTaskRepository(db)
	ctx := context.Background()

	taskID := uuid.NewString()
	
	// Insert initial task
	task := &domain.Task{
		ID:             taskID,
		Name:           "Test Task 3",
		CronExpression: "* * * * *",
		Payload: domain.HTTPRequestInfo{
			URL:    "http://example.com",
			Method: "GET",
		},
		Status:    domain.TaskStatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Version:   1,
	}

	err := repo.Save(ctx, task)
	require.NoError(t, err)

	// Try to save with the same version (should fail)
	task2 := &domain.Task{
		ID:             taskID,
		Name:           "Conflicting Task",
		CronExpression: "* * * * *",
		Payload: domain.HTTPRequestInfo{
			URL:    "http://example.com",
			Method: "GET",
		},
		Status:    domain.TaskStatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Version:   1, // Same version, should conflict
	}

	err = repo.Save(ctx, task2)
	assert.ErrorIs(t, err, domain.ErrConflict)

	// Save with correct version should succeed
	task3 := &domain.Task{
		ID:             taskID,
		Name:           "Correct Version Task",
		CronExpression: "* * * * *",
		Payload: domain.HTTPRequestInfo{
			URL:    "http://example.com",
			Method: "GET",
		},
		Status:    domain.TaskStatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Version:   2,
	}

	err = repo.Save(ctx, task3)
	assert.NoError(t, err)
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
		Version:   1,
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
		Version:   1,
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
		Version:       1,
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
