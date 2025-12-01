package memory

import (
	"context"
	"sync"

	"github.com/yourname/go-dist-scheduler/internal/domain"
)

type InMemoryTaskRepository struct {
	mu    sync.Mutex
	tasks map[string]*domain.Task
}

func NewInMemoryTaskRepository() *InMemoryTaskRepository {
	return &InMemoryTaskRepository{
		tasks: make(map[string]*domain.Task),
	}
}

func (r *InMemoryTaskRepository) Save(ctx context.Context, task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[task.ID] = task
	return nil
}

func (r *InMemoryTaskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if task, ok := r.tasks[id]; ok {
		return task, nil
	}
	return nil, nil
}

func (r *InMemoryTaskRepository) FindAllActive(ctx context.Context) ([]*domain.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var activeTasks []*domain.Task
	for _, task := range r.tasks {
		if task.Status == domain.TaskStatusActive {
			activeTasks = append(activeTasks, task)
		}
	}
	return activeTasks, nil
}
