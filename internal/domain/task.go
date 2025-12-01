package domain

import (
	"time"

	"github.com/robfig/cron/v3"
)

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

func (t *Task) NextRunTime(now time.Time) (time.Time, error) {
	specParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := specParser.Parse(t.CronExpression)
	if err != nil {
		return time.Time{}, err
	}
	return schedule.Next(now), nil
}
