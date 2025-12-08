package domain

import "time"

// JobID は Job の一意な識別子を表す型です。
// string をラップすることで型安全性を提供します。
type JobID string

type JobStatus int

const (
	JobStatusPending JobStatus = iota
	JobStatusRunning
	JobStatusSuccess
	JobStatusFailed
)

type Job struct {
	ID          JobID
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
