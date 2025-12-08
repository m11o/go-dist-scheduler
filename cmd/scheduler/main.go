package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/yourname/go-dist-scheduler/internal/domain"
	"github.com/yourname/go-dist-scheduler/internal/infrastructure/memory"
	"github.com/yourname/go-dist-scheduler/internal/usecase"
)

func main() {
	log.Println("Starting go-dist-scheduler...")

	// インメモリリポジトリの初期化
	taskRepo := memory.NewInMemoryTaskRepository()
	jobRepo := memory.NewInMemoryJobRepository()

	// ユースケースの初期化（DI）
	scheduler := usecase.NewScheduler(taskRepo, jobRepo)
	executor := usecase.NewExecutor(jobRepo)

	ctx := context.Background()

	// サンプルタスクの登録（1分ごとに実行）
	// 注: このタスクはデモンストレーション用です。実際のHTTPリクエストは送信されません。
	sampleTask := &domain.Task{
		ID:             uuid.New().String(),
		Name:           "Sample Task",
		CronExpression: "* * * * *", // 1分ごと（分・時・日・月・曜日の5フィールド形式）
		Payload: domain.HTTPRequestInfo{
			URL:    "http://example.com/webhook",
			Method: "POST",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: []byte(`{"message":"Hello from scheduler"}`),
		},
		Status:    domain.TaskStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   0,
	}

	if err := taskRepo.Save(ctx, sampleTask); err != nil {
		log.Fatalf("Failed to save sample task: %v", err)
	}
	log.Printf("Registered sample task: %s (ID: %s)", sampleTask.Name, sampleTask.ID)

	// 1秒ごとにスケジューラーとエグゼキューターを実行
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// シグナルハンドリングによるグレースフルシャットダウン
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	log.Println("Scheduler loop started. Checking every 1 second...")
	log.Println("Press Ctrl+C to stop")

	for {
		select {
		case <-ticker.C:
			now := time.Now()

			// スケジューラー: タスクをチェックしてジョブをエンキュー
			if err := scheduler.CheckAndEnqueue(ctx, now); err != nil {
				log.Printf("Error in CheckAndEnqueue: %v", err)
			}

			// エグゼキューター: ペンディング中のジョブを実行
			if err := executor.RunPendingJob(ctx); err != nil {
				log.Printf("Error in RunPendingJob: %v", err)
			}
		case sig := <-sigCh:
			log.Printf("Received signal: %v. Shutting down gracefully...", sig)
			return
		}
	}
}
