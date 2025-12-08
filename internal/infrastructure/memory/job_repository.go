package memory

import (
	"context"
	"sync"

	"github.com/yourname/go-dist-scheduler/internal/domain"
)

// InMemoryJobRepository implements domain.JobRepository for in-memory job persistence.
type InMemoryJobRepository struct {
	mu   sync.Mutex
	jobs map[string]*domain.Job
}

func NewInMemoryJobRepository() *InMemoryJobRepository {
	return &InMemoryJobRepository{
		jobs: make(map[string]*domain.Job),
	}
}

func (r *InMemoryJobRepository) Save(ctx context.Context, job *domain.Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs[job.ID] = copyJob(job)
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

func (r *InMemoryJobRepository) Update(ctx context.Context, job *domain.Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// 更新がアトミックであり、渡されたジョブの完全な状態を反映することを保証するため、
	// フィールドを個別に更新するのではなく、マップ内のオブジェクトを置き換えます。
	if _, ok := r.jobs[job.ID]; ok {
		r.jobs[job.ID] = copyJob(job)
	}
	return nil
}

// InMemoryJobQueue implements domain.JobQueue for in-memory job queueing.
type InMemoryJobQueue struct {
	mu    sync.Mutex
	queue []string
	jobs  map[string]*domain.Job
}

func NewInMemoryJobQueue() *InMemoryJobQueue {
	return &InMemoryJobQueue{
		queue: make([]string, 0),
		jobs:  make(map[string]*domain.Job),
	}
}

func (q *InMemoryJobQueue) Enqueue(ctx context.Context, job *domain.Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.jobs[job.ID] = copyJob(job)
	q.queue = append(q.queue, job.ID)
	return nil
}

func (q *InMemoryJobQueue) Dequeue(ctx context.Context) (*domain.Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.queue) == 0 {
		return nil, nil
	}
	jobID := q.queue[0]
	q.queue = q.queue[1:]
	return copyJob(q.jobs[jobID]), nil
}

// copyJob creates a shallow copy of a Job object.
func copyJob(j *domain.Job) *domain.Job {
	if j == nil {
		return nil
	}
	c := *j
	return &c
}
