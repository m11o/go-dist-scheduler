package domain

import "time"

type TaskStatus int

const (
	TaskStatusActive TaskStatus = iota
	TaskStatusPaused
)

type HTTPRequestInfo struct {
	URL     string
	Method  string
	Headers map[string]string
	Body    []byte
}

type Task struct {
	ID             string
	Name           string
	CronExpression string
	Payload        HTTPRequestInfo
	Status         TaskStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
