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
	LastCheckedAt  time.Time
}

// CronParser は、標準的な5フィールド（分・時・日・月・曜日）のCron式を解析するパーサーです。
// このパーサーはパッケージレベルで一度だけ生成され、複数のgoroutineから安全に利用できます。
var CronParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

func (t *Task) NextRunTime(now time.Time) (time.Time, error) {
	schedule, err := CronParser.Parse(t.CronExpression)
	if err != nil {
		return time.Time{}, err
	}
	return schedule.Next(now), nil
}
