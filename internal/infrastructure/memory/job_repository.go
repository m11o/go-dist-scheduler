package memory

import (
	"context"
	"sync"

	"github.com/yourname/go-dist-scheduler/internal/domain"
)

type InMemoryJobRepository struct {
	mu      sync.Mutex
	jobs    map[string]*domain.Job
	queue   []string
}

func NewInMemoryJobRepository() *InMemoryJobRepository {
	return &InMemoryJobRepository{
		jobs:  make(map[string]*domain.Job),
		queue: make([]string, 0),
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

func (r *InMemoryJobRepository) UpdateStatus(ctx context.Context, job *domain.Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// 更新がアトミックであり、渡されたジョブの完全な状態を反映することを保証するため、
	// フィールドを個別に更新するのではなく、マップ内のオブジェクトを置き換えます。
	if _, ok := r.jobs[job.ID]; ok {
		r.jobs[job.ID] = copyJob(job)
	}
	return nil
}

func (r *InMemoryJobRepository) FindByID(ctx context.Context, id string) (*domain.Job, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if job, ok := r.jobs[id]; ok {
		return copyJob(job), nil
	}
	return nil, nil
}

// copyJob creates a shallow copy of a Job object.
func copyJob(j *domain.Job) *domain.Job {
	if j == nil {
		return nil
	}
	c := *j
	return &c
}
