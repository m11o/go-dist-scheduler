package postgres

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/yourname/go-dist-scheduler/internal/domain"
)

// TaskDTO represents the database row structure for a Task.
type TaskDTO struct {
	ID             string         `db:"id"`
	Name           string         `db:"name"`
	CronExpression string         `db:"cron_expression"`
	Payload        []byte         `db:"payload"`
	Status         int            `db:"status"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
	LastCheckedAt  sql.NullTime   `db:"last_checked_at"`
}

// payloadJSON represents the JSON structure stored in the payload column.
type payloadJSON struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"` // base64-encoded
}

// ToDTO converts a domain Task to a TaskDTO.
func ToDTO(task *domain.Task) (*TaskDTO, error) {
	// Convert HTTPRequestInfo to JSON
	payload := payloadJSON{
		URL:     task.Payload.URL,
		Method:  task.Payload.Method,
		Headers: task.Payload.Headers,
		Body:    base64.StdEncoding.EncodeToString(task.Payload.Body),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	dto := &TaskDTO{
		ID:             task.ID,
		Name:           task.Name,
		CronExpression: task.CronExpression,
		Payload:        payloadBytes,
		Status:         int(task.Status),
		CreatedAt:      task.CreatedAt,
		UpdatedAt:      task.UpdatedAt,
	}

	if !task.LastCheckedAt.IsZero() {
		dto.LastCheckedAt = sql.NullTime{Time: task.LastCheckedAt, Valid: true}
	}

	return dto, nil
}

// ToDomain converts a TaskDTO to a domain Task.
func (dto *TaskDTO) ToDomain() (*domain.Task, error) {
	// Parse JSON payload
	var payload payloadJSON
	if err := json.Unmarshal(dto.Payload, &payload); err != nil {
		return nil, err
	}

	// Decode base64 body
	var body []byte
	if payload.Body != "" {
		var err error
		body, err = base64.StdEncoding.DecodeString(payload.Body)
		if err != nil {
			return nil, err
		}
	}

	task := &domain.Task{
		ID:             dto.ID,
		Name:           dto.Name,
		CronExpression: dto.CronExpression,
		Payload: domain.HTTPRequestInfo{
			URL:     payload.URL,
			Method:  payload.Method,
			Headers: payload.Headers,
			Body:    body,
		},
		Status:    domain.TaskStatus(dto.Status),
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
	}

	if dto.LastCheckedAt.Valid {
		task.LastCheckedAt = dto.LastCheckedAt.Time
	}

	return task, nil
}
