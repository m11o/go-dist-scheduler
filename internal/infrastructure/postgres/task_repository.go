package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/yourname/go-dist-scheduler/internal/domain"
)

// TaskRepository is a PostgreSQL implementation of the TaskRepository interface.
type TaskRepository struct {
	db *sql.DB
}

// NewTaskRepository creates a new PostgreSQL TaskRepository.
func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// Save saves a task to the database.
func (r *TaskRepository) Save(ctx context.Context, task *domain.Task) error {
	dto, err := ToDTO(task)
	if err != nil {
		return fmt.Errorf("failed to convert task to DTO: %w", err)
	}

	// Check if task exists
	var existingVersion sql.NullInt64
	err = r.db.QueryRowContext(ctx, "SELECT version FROM tasks WHERE id = $1", dto.ID).Scan(&existingVersion)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to check existing task: %w", err)
	}

	// Check for version conflict (optimistic locking)
	if existingVersion.Valid && int(existingVersion.Int64) != dto.Version-1 {
		return domain.ErrConflict
	}

	var result sql.Result
	if existingVersion.Valid {
		// Update existing task with version check
		query := `
			UPDATE tasks
			SET name = $2, cron_expression = $3, payload = $4, status = $5,
				updated_at = $6, last_checked_at = $7, version = $8
			WHERE id = $1 AND version = $9
		`
		result, err = r.db.ExecContext(ctx, query,
			dto.ID,
			dto.Name,
			dto.CronExpression,
			dto.Payload,
			dto.Status,
			dto.UpdatedAt,
			dto.LastCheckedAt,
			dto.Version,
			dto.Version-1, // Expected previous version
		)
	} else {
		// Insert new task
		query := `
			INSERT INTO tasks (id, name, cron_expression, payload, status, created_at, updated_at, last_checked_at, version)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		result, err = r.db.ExecContext(ctx, query,
			dto.ID,
			dto.Name,
			dto.CronExpression,
			dto.Payload,
			dto.Status,
			dto.CreatedAt,
			dto.UpdatedAt,
			dto.LastCheckedAt,
			dto.Version,
		)
	}

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Check for unique violation or other constraint violations
			if pqErr.Code == "23505" { // unique_violation
				return domain.ErrConflict
			}
		}
		return fmt.Errorf("failed to save task: %w", err)
	}

	// Check if update affected any rows (for optimistic locking)
	if existingVersion.Valid {
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}
		if rowsAffected == 0 {
			return domain.ErrConflict
		}
	}

	return nil
}

// FindByID finds a task by its ID.
func (r *TaskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
	query := `
		SELECT id, name, cron_expression, payload, status, created_at, updated_at, last_checked_at, version
		FROM tasks
		WHERE id = $1
	`

	var dto TaskDTO
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&dto.ID,
		&dto.Name,
		&dto.CronExpression,
		&dto.Payload,
		&dto.Status,
		&dto.CreatedAt,
		&dto.UpdatedAt,
		&dto.LastCheckedAt,
		&dto.Version,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}

	task, err := dto.ToDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to convert DTO to domain: %w", err)
	}

	return task, nil
}

// FindAllActive finds all active tasks.
func (r *TaskRepository) FindAllActive(ctx context.Context) ([]*domain.Task, error) {
	query := `
		SELECT id, name, cron_expression, payload, status, created_at, updated_at, last_checked_at, version
		FROM tasks
		WHERE status = $1
	`

	rows, err := r.db.QueryContext(ctx, query, int(domain.TaskStatusActive))
	if err != nil {
		return nil, fmt.Errorf("failed to query active tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		var dto TaskDTO
		err := rows.Scan(
			&dto.ID,
			&dto.Name,
			&dto.CronExpression,
			&dto.Payload,
			&dto.Status,
			&dto.CreatedAt,
			&dto.UpdatedAt,
			&dto.LastCheckedAt,
			&dto.Version,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task row: %w", err)
		}

		task, err := dto.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert DTO to domain: %w", err)
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task rows: %w", err)
	}

	return tasks, nil
}
