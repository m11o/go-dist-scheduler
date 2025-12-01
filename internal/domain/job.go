package domain

import "time"

type JobStatus int

const (
	JobStatusPending JobStatus = iota
	JobStatusRunning
	JobStatusSuccess
	JobStatusFailed
)

type Job struct {
	ID          string
	TaskID      string
	ScheduledAt time.Time
	StartedAt   time.Time
	FinishedAt  time.Time
	Status      JobStatus
	RetryCount  int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
