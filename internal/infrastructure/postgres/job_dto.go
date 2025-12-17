package postgres

import (
	"database/sql"
	"time"

	"github.com/yourname/go-dist-scheduler/internal/domain"
)

// JobDTO represents the database row structure for a Job.
type JobDTO struct {
	ID          string       `db:"id"`
	TaskID      string       `db:"task_id"`
	ScheduledAt time.Time    `db:"scheduled_at"`
	StartedAt   sql.NullTime `db:"started_at"`
	FinishedAt  sql.NullTime `db:"finished_at"`
	Status      int          `db:"status"`
	RetryCount  int          `db:"retry_count"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
}

// ToJobDTO converts a domain Job to a JobDTO.
func ToJobDTO(job *domain.Job) *JobDTO {
	dto := &JobDTO{
		ID:          job.ID,
		TaskID:      job.TaskID,
		ScheduledAt: job.ScheduledAt,
		Status:      int(job.Status),
		RetryCount:  job.RetryCount,
		CreatedAt:   job.CreatedAt,
		UpdatedAt:   job.UpdatedAt,
	}

	if !job.StartedAt.IsZero() {
		dto.StartedAt = sql.NullTime{Time: job.StartedAt, Valid: true}
	}

	if !job.FinishedAt.IsZero() {
		dto.FinishedAt = sql.NullTime{Time: job.FinishedAt, Valid: true}
	}

	return dto
}

// ToDomain converts a JobDTO to a domain Job.
func (dto *JobDTO) ToJobDomain() *domain.Job {
	job := &domain.Job{
		ID:          dto.ID,
		TaskID:      dto.TaskID,
		ScheduledAt: dto.ScheduledAt,
		Status:      domain.JobStatus(dto.Status),
		RetryCount:  dto.RetryCount,
		CreatedAt:   dto.CreatedAt,
		UpdatedAt:   dto.UpdatedAt,
	}

	if dto.StartedAt.Valid {
		job.StartedAt = dto.StartedAt.Time
	}

	if dto.FinishedAt.Valid {
		job.FinishedAt = dto.FinishedAt.Time
	}

	return job
}
