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

func (j *Job) MarkAsRunning() {
	j.Status = JobStatusRunning
	j.StartedAt = time.Now()
	j.UpdatedAt = time.Now()
}

func (j *Job) MarkAsSuccess() {
	j.Status = JobStatusSuccess
	j.FinishedAt = time.Now()
	j.UpdatedAt = time.Now()
}

func (j *Job) MarkAsFailed() {
	j.Status = JobStatusFailed
	j.FinishedAt = time.Now()
	j.UpdatedAt = time.Now()
}
