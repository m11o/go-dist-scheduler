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

// Save saves a task to the database using pessimistic locking.
func (r *TaskRepository) Save(ctx context.Context, task *domain.Task) error {
	dto, err := ToDTO(task)
	if err != nil {
		return fmt.Errorf("failed to convert task to DTO: %w", err)
	}

	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // Rollback is safe to call even after Commit
	}()

	// Use SELECT ... FOR UPDATE to acquire pessimistic lock on the row
	var lockedID sql.NullString
	err = tx.QueryRowContext(ctx, "SELECT id FROM tasks WHERE id = $1 FOR UPDATE", dto.ID).Scan(&lockedID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to lock task: %w", err)
	}

	if lockedID.Valid {
		// Update existing task
		query := `
			UPDATE tasks
			SET name = $2, cron_expression = $3, payload = $4, status = $5,
				updated_at = $6, last_checked_at = $7
			WHERE id = $1
		`
		_, err = tx.ExecContext(ctx, query,
			dto.ID,
			dto.Name,
			dto.CronExpression,
			dto.Payload,
			dto.Status,
			dto.UpdatedAt,
			dto.LastCheckedAt,
		)
	} else {
		// Insert new task
		query := `
			INSERT INTO tasks (id, name, cron_expression, payload, status, created_at, updated_at, last_checked_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err = tx.ExecContext(ctx, query,
			dto.ID,
			dto.Name,
			dto.CronExpression,
			dto.Payload,
			dto.Status,
			dto.CreatedAt,
			dto.UpdatedAt,
			dto.LastCheckedAt,
		)
	}

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Check for unique violation or other constraint violations
			if pqErr.Code == "23505" { // unique_violation
				return domain.ErrConstraintViolation
			}
		}
		return fmt.Errorf("failed to save task: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FindByID finds a task by its ID.
func (r *TaskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
	query := `
		SELECT id, name, cron_expression, payload, status, created_at, updated_at, last_checked_at
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
		SELECT id, name, cron_expression, payload, status, created_at, updated_at, last_checked_at
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
