package domain

import "context"

type TaskRepository interface {
	Save(ctx context.Context, task *Task) error
	FindByID(ctx context.Context, id string) (*Task, error)
	FindAllActive(ctx context.Context) ([]*Task, error)
}

type JobRepository interface {
	Enqueue(ctx context.Context, job *Job) error
	Dequeue(ctx context.Context) (*Job, error)
	UpdateStatus(ctx context.Context, job *Job) error
}
