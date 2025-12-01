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
	r.tasks[task.ID] = copyTask(task)
	return nil
}

func (r *InMemoryTaskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if task, ok := r.tasks[id]; ok {
		return copyTask(task), nil
	}
	return nil, nil
}

func (r *InMemoryTaskRepository) FindAllActive(ctx context.Context) ([]*domain.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var activeTasks []*domain.Task
	for _, task := range r.tasks {
		if task.Status == domain.TaskStatusActive {
			activeTasks = append(activeTasks, copyTask(task))
		}
	}
	return activeTasks, nil
}

// copyTask creates a deep copy of a Task object.
func copyTask(t *domain.Task) *domain.Task {
	if t == nil {
		return nil
	}

	c := *t // Shallow copy of the struct

	// Deep copy the Headers map
	if t.Payload.Headers != nil {
		c.Payload.Headers = make(map[string]string, len(t.Payload.Headers))
		for k, v := range t.Payload.Headers {
			c.Payload.Headers[k] = v
		}
	}

	// Deep copy the Body slice
	if t.Payload.Body != nil {
		c.Payload.Body = make([]byte, len(t.Payload.Body))
		copy(c.Payload.Body, t.Payload.Body)
	}

	return &c
}
