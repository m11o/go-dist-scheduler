package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/yourname/go-dist-scheduler/internal/domain"
)

// JobRepository is a PostgreSQL implementation of the JobRepository interface.
type JobRepository struct {
	db *sql.DB
}

// NewJobRepository creates a new PostgreSQL JobRepository.
func NewJobRepository(db *sql.DB) *JobRepository {
	return &JobRepository{db: db}
}

// Enqueue adds a job to the queue.
func (r *JobRepository) Enqueue(ctx context.Context, job *domain.Job) error {
	dto := ToJobDTO(job)

	query := `
		INSERT INTO jobs (id, task_id, scheduled_at, started_at, finished_at, status, retry_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		dto.ID,
		dto.TaskID,
		dto.ScheduledAt,
		dto.StartedAt,
		dto.FinishedAt,
		dto.Status,
		dto.RetryCount,
		dto.CreatedAt,
		dto.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	return nil
}

// Dequeue retrieves and returns the next pending job from the queue.
// It uses SELECT ... FOR UPDATE SKIP LOCKED to ensure concurrent workers
// don't pick up the same job.
func (r *JobRepository) Dequeue(ctx context.Context) (*domain.Job, error) {
	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // Rollback is safe to call even after Commit
	}()

	// Select the first pending job with pessimistic locking
	query := `
		SELECT id, task_id, scheduled_at, started_at, finished_at, status, retry_count, created_at, updated_at
		FROM jobs
		WHERE status = $1
		ORDER BY scheduled_at ASC, created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`

	var dto JobDTO
	err = tx.QueryRowContext(ctx, query, int(domain.JobStatusPending)).Scan(
		&dto.ID,
		&dto.TaskID,
		&dto.ScheduledAt,
		&dto.StartedAt,
		&dto.FinishedAt,
		&dto.Status,
		&dto.RetryCount,
		&dto.CreatedAt,
		&dto.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	// Commit transaction to release the lock
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return dto.ToJobDomain(), nil
}

// UpdateStatus updates the status of a job.
func (r *JobRepository) UpdateStatus(ctx context.Context, jobID string, status domain.JobStatus) error {
	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // Rollback is safe to call even after Commit
	}()

	// Lock the job row
	var lockedID sql.NullString
	err = tx.QueryRowContext(ctx, "SELECT id FROM jobs WHERE id = $1 FOR UPDATE", jobID).Scan(&lockedID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("job not found: %s", jobID)
	}
	if err != nil {
		return fmt.Errorf("failed to lock job: %w", err)
	}

	// Retrieve current job data
	query := `
		SELECT id, task_id, scheduled_at, started_at, finished_at, status, retry_count, created_at, updated_at
		FROM jobs
		WHERE id = $1
	`
	var dto JobDTO
	err = tx.QueryRowContext(ctx, query, jobID).Scan(
		&dto.ID,
		&dto.TaskID,
		&dto.ScheduledAt,
		&dto.StartedAt,
		&dto.FinishedAt,
		&dto.Status,
		&dto.RetryCount,
		&dto.CreatedAt,
		&dto.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to retrieve job: %w", err)
	}

	// Convert to domain and apply status change
	job := dto.ToJobDomain()
	switch status {
	case domain.JobStatusRunning:
		job.MarkAsRunning()
	case domain.JobStatusSuccess:
		job.MarkAsSuccess()
	case domain.JobStatusFailed:
		job.MarkAsFailed()
	default:
		job.Status = status
		job.UpdatedAt = time.Now()
	}

	// Convert back to DTO
	updatedDTO := ToJobDTO(job)

	// Update job in database
	updateQuery := `
		UPDATE jobs
		SET status = $2, started_at = $3, finished_at = $4, updated_at = $5
		WHERE id = $1
	`
	_, err = tx.ExecContext(ctx, updateQuery,
		updatedDTO.ID,
		updatedDTO.Status,
		updatedDTO.StartedAt,
		updatedDTO.FinishedAt,
		updatedDTO.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
