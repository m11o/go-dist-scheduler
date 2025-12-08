package domain

import "context"

type TaskRepository interface {
	Save(ctx context.Context, task *Task) error
	FindByID(ctx context.Context, id string) (*Task, error)
	FindAllActive(ctx context.Context) ([]*Task, error)
}

type JobRepository interface {
	Save(ctx context.Context, job *Job) error
	FindByID(ctx context.Context, id string) (*Job, error)
	Update(ctx context.Context, job *Job) error
}

type JobQueue interface {
	Enqueue(ctx context.Context, job *Job) error
	Dequeue(ctx context.Context) (*Job, error)
}
