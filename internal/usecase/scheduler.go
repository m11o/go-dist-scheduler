package usecase

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/yourname/go-dist-scheduler/internal/domain"
)

// Scheduler は、タスクをチェックしてジョブをエンキューするユースケースを担当します。
type Scheduler struct {
	taskRepo domain.TaskRepository
	jobRepo  domain.JobRepository
}

// NewScheduler は新しいSchedulerインスタンスを生成します。
func NewScheduler(taskRepo domain.TaskRepository, jobRepo domain.JobRepository) *Scheduler {
	return &Scheduler{
		taskRepo: taskRepo,
		jobRepo:  jobRepo,
	}
}

// CheckAndEnqueue は、実行時刻が到来したタスクを元にジョブを作成し、キューに追加します。
// スケジューラのダウンタイムなどで実行されなかったジョブも、遅れてエンキューされます。
func (s *Scheduler) CheckAndEnqueue(ctx context.Context, now time.Time) error {
	tasks, err := s.taskRepo.FindAllActive(ctx)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		// cron式をパース
		schedule, err := domain.CronParser.Parse(task.CronExpression)
		if err != nil {
			log.Printf("failed to parse cron expression for task %s: %v", task.ID, err)
			continue
		}

		// 前回チェックした時刻を取得。なければタスクの作成時刻を仕様
		lastChecked := task.LastCheckedAt
		if lastChecked.IsZero() {
			lastChecked = task.CreatedAt
		}

		// 前回チェック時から現在時刻までの間に実行されるべきだった時刻をすべて取得
		nextRunTime := schedule.Next(lastChecked)
		for !nextRunTime.IsZero() && !nextRunTime.After(now) {
			newJob := &domain.Job{
				ID:          uuid.New().String(),
				TaskID:      task.ID,
				ScheduledAt: nextRunTime,
				Status:      domain.JobStatusPending,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			if err := s.jobRepo.Enqueue(ctx, newJob); err != nil {
				log.Printf("failed to enqueue job for task %s: %v", task.ID, err)
				// 1件でもエンキューに失敗したら、このタスクの処理は中断
				// lastChecked を更新しないことで、次回再実行されるようにする
				goto nextTask
			}
			// 次の実行時刻を計算
			nextRunTime = schedule.Next(nextRunTime)
		}

		// 最終チェック時刻を更新
		task.LastCheckedAt = now
		if err := s.taskRepo.Save(ctx, task); err != nil {
			log.Printf("failed to update last checked time for task %s: %v", task.ID, err)
		}
	nextTask:
	}

	return nil
}
