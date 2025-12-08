package usecase

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yourname/go-dist-scheduler/internal/domain"
)

// mockTaskRepository は TaskRepository のモック実装です。
type mockTaskRepository struct {
	mu               sync.Mutex
	tasks            map[string]*domain.Task
	findAllActiveErr error
}

func (m *mockTaskRepository) Save(ctx context.Context, task *domain.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, ok := m.tasks[task.ID]
	if ok && existing.Version != task.Version-1 {
		return domain.ErrConflict
	}
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.tasks[id], nil
}

func (m *mockTaskRepository) FindAllActive(ctx context.Context) ([]*domain.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.findAllActiveErr != nil {
		return nil, m.findAllActiveErr
	}
	var activeTasks []*domain.Task
	for _, task := range m.tasks {
		if task.Status == domain.TaskStatusActive {
			activeTasks = append(activeTasks, task)
		}
	}
	return activeTasks, nil
}

// mockJobRepository は JobRepository のモック実装です。
type mockJobRepository struct {
	mu      sync.Mutex
	jobs    map[domain.JobID]*domain.Job
	saveErr error
}

func (m *mockJobRepository) Save(ctx context.Context, job *domain.Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.saveErr != nil {
		return m.saveErr
	}
	if m.jobs == nil {
		m.jobs = make(map[domain.JobID]*domain.Job)
	}
	m.jobs[job.ID] = job
	return nil
}

func (m *mockJobRepository) FindByID(ctx context.Context, id domain.JobID) (*domain.Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.jobs[id], nil
}

func (m *mockJobRepository) Update(ctx context.Context, job *domain.Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.jobs[job.ID]; ok {
		m.jobs[job.ID] = job
	}
	return nil
}

// mockJobQueue は JobQueue のモック実装です。
type mockJobQueue struct {
	mu         sync.Mutex
	enqueued   []domain.JobID
	enqueueErr error
}

func (m *mockJobQueue) Enqueue(ctx context.Context, jobID domain.JobID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.enqueueErr != nil {
		return m.enqueueErr
	}
	m.enqueued = append(m.enqueued, jobID)
	return nil
}

func (m *mockJobQueue) Dequeue(ctx context.Context) (domain.JobID, error) {
	return "", nil
}

func TestScheduler_CheckAndEnqueue(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	now := time.Date(2023, 10, 28, 10, 0, 0, 0, jst)

	testCases := []struct {
		name                string
		tasks               map[string]*domain.Task
		now                 time.Time
		expectedJobs        int
		expectedLastChecked map[string]time.Time
	}{
		{
			name: "should enqueue jobs since creation if never checked (2 jobs)",
			tasks: map[string]*domain.Task{
				"task1": {ID: "task1", CronExpression: "* * * * *", Status: domain.TaskStatusActive, CreatedAt: now.Add(-2 * time.Minute)},
			},
			now:                 now,
			expectedJobs:        2,
			expectedLastChecked: map[string]time.Time{"task1": now},
		},
		{
			name: "should not enqueue a job for a task that is not due",
			tasks: map[string]*domain.Task{
				"task2": {ID: "task2", CronExpression: "1 * * * *", Status: domain.TaskStatusActive, CreatedAt: now.Add(-2 * time.Minute)},
			},
			now:                 now,
			expectedJobs:        0,
			expectedLastChecked: map[string]time.Time{"task2": now},
		},
		{
			name: "should enqueue missed jobs since last check",
			tasks: map[string]*domain.Task{
				"task3": {ID: "task3", CronExpression: "* * * * *", Status: domain.TaskStatusActive, CreatedAt: now.Add(-10 * time.Minute), LastCheckedAt: now.Add(-5 * time.Minute)},
			},
			now:                 now,
			expectedJobs:        5,
			expectedLastChecked: map[string]time.Time{"task3": now},
		},
		{
			name: "should enqueue jobs since creation if never checked",
			tasks: map[string]*domain.Task{
				"task4": {ID: "task4", CronExpression: "*/2 * * * *", Status: domain.TaskStatusActive, CreatedAt: now.Add(-10 * time.Minute)},
			},
			now:                 now,
			expectedJobs:        5,
			expectedLastChecked: map[string]time.Time{"task4": now},
		},
		{
			name: "should not enqueue jobs for paused tasks",
			tasks: map[string]*domain.Task{
				"task5": {ID: "task5", CronExpression: "* * * * *", Status: domain.TaskStatusPaused, CreatedAt: now.Add(-10 * time.Minute)},
			},
			now:                 now,
			expectedJobs:        0,
			expectedLastChecked: map[string]time.Time{}, // LastCheckedAt should not be updated for paused tasks
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Deep copy tasks to avoid race conditions in tests
			tasksCopy := make(map[string]*domain.Task)
			for k, v := range tc.tasks {
				taskCopy := *v
				tasksCopy[k] = &taskCopy
			}
			taskRepo := &mockTaskRepository{tasks: tasksCopy}
			jobRepo := &mockJobRepository{}
			jobQueue := &mockJobQueue{}
			scheduler := NewScheduler(taskRepo, jobRepo, jobQueue)

			err := scheduler.CheckAndEnqueue(context.Background(), tc.now)
			assert.NoError(t, err)
			assert.Len(t, jobQueue.enqueued, tc.expectedJobs)

			for taskID, expectedTime := range tc.expectedLastChecked {
				task, err := taskRepo.FindByID(context.Background(), taskID)
				assert.NoError(t, err)
				assert.NotNil(t, task)
				assert.Equal(t, expectedTime, task.LastCheckedAt)
			}
		})
	}

	t.Run("should return error when FindAllActive fails", func(t *testing.T) {
		taskRepo := &mockTaskRepository{findAllActiveErr: assert.AnError}
		jobRepo := &mockJobRepository{}
		jobQueue := &mockJobQueue{}
		scheduler := NewScheduler(taskRepo, jobRepo, jobQueue)

		err := scheduler.CheckAndEnqueue(context.Background(), now)
		assert.Error(t, err)
	})

	t.Run("should not return error when Enqueue fails", func(t *testing.T) {
		tasks := map[string]*domain.Task{
			"task1": {ID: "task1", CronExpression: "* * * * *", Status: domain.TaskStatusActive, CreatedAt: now.Add(-2 * time.Minute)},
		}
		taskRepo := &mockTaskRepository{tasks: tasks}
		jobRepo := &mockJobRepository{}
		jobQueue := &mockJobQueue{enqueueErr: assert.AnError}
		scheduler := NewScheduler(taskRepo, jobRepo, jobQueue)

		err := scheduler.CheckAndEnqueue(context.Background(), now)
		assert.NoError(t, err)
	})
}

func TestScheduler_CheckAndEnqueue_Conflict(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	now := time.Date(2023, 10, 28, 10, 0, 0, 0, jst)

	task := &domain.Task{ID: "task1", CronExpression: "* * * * *", Status: domain.TaskStatusActive, CreatedAt: now.Add(-2 * time.Minute), Version: 1}
	tasks := map[string]*domain.Task{"task1": task}

	taskRepo := &mockTaskRepository{tasks: tasks}
	jobRepo := &mockJobRepository{}
	jobQueue := &mockJobQueue{}
	scheduler1 := NewScheduler(taskRepo, jobRepo, jobQueue)

	// Simulate scheduler1 running first and updating the task
	err1 := scheduler1.CheckAndEnqueue(context.Background(), now)
	assert.NoError(t, err1)
	assert.Len(t, jobQueue.enqueued, 2)

	// Create a new scheduler with an outdated task to simulate a race condition
	taskCopy := *task
	outdatedTasks := map[string]*domain.Task{"task1": &taskCopy}
	scheduler2 := NewScheduler(&mockTaskRepository{tasks: outdatedTasks}, jobRepo, jobQueue)

	// Simulate scheduler2 running concurrently with an outdated task version
	err2 := scheduler2.CheckAndEnqueue(context.Background(), now)
	assert.NoError(t, err2)
	assert.Len(t, jobQueue.enqueued, 2) // No new jobs should be enqueued

	finalTask, err := taskRepo.FindByID(context.Background(), "task1")
	assert.NoError(t, err)
	assert.Equal(t, 2, finalTask.Version)
}
