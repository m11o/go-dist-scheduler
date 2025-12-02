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
		lastChecked := task.LastCheckedAt
		if lastChecked.IsZero() {
			lastChecked = task.CreatedAt
		}

		dueRunTimes, err := task.GetDueRunTimes(lastChecked, now)
		if err != nil {
			log.Printf("failed to get due run times for task %s: %v", task.ID, err)
			continue
		}

		for _, runTime := range dueRunTimes {
			newJob := &domain.Job{
				ID:          uuid.New().String(),
				TaskID:      task.ID,
				ScheduledAt: runTime,
				Status:      domain.JobStatusPending,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			if err := s.jobRepo.Enqueue(ctx, newJob); err != nil {
				log.Printf("failed to enqueue job for task %s: %v", task.ID, err)
				goto nextTask
			}
		}

		task.LastCheckedAt = now
		task.Version++
		if err := s.taskRepo.Save(ctx, task); err != nil {
			if err == domain.ErrConflict {
				log.Printf("conflict updating task %s, skipping", task.ID)
			} else {
				log.Printf("failed to update last checked time for task %s: %v", task.ID, err)
			}
		}
	nextTask:
	}

	return nil
}
