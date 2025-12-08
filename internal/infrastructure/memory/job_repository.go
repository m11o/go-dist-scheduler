package memory

import (
	"context"
	"sync"

	"github.com/yourname/go-dist-scheduler/internal/domain"
)

// InMemoryJobRepository implements domain.JobRepository for in-memory job queueing.
type InMemoryJobRepository struct {
	mu    sync.Mutex
	queue []string
	jobs  map[string]*domain.Job
}

func NewInMemoryJobRepository() *InMemoryJobRepository {
	return &InMemoryJobRepository{
		queue: make([]string, 0),
		jobs:  make(map[string]*domain.Job),
	}
}

func (r *InMemoryJobRepository) Enqueue(ctx context.Context, job *domain.Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs[job.ID] = copyJob(job)
	r.queue = append(r.queue, job.ID)
	return nil
}

func (r *InMemoryJobRepository) Dequeue(ctx context.Context) (*domain.Job, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.queue) == 0 {
		return nil, nil
	}
	jobID := r.queue[0]
	r.queue = r.queue[1:]
	return copyJob(r.jobs[jobID]), nil
}

func (r *InMemoryJobRepository) UpdateStatus(ctx context.Context, jobID string, status domain.JobStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if job, ok := r.jobs[jobID]; ok {
		switch status {
		case domain.JobStatusRunning:
			job.MarkAsRunning()
		case domain.JobStatusSuccess:
			job.MarkAsSuccess()
		case domain.JobStatusFailed:
			job.MarkAsFailed()
		}
	}
	return nil
}

// copyJob creates a shallow copy of a Job object.
func copyJob(j *domain.Job) *domain.Job {
	if j == nil {
		return nil
	}
	c := *j
	return &c
}
