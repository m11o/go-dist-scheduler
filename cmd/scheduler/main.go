package main

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/yourname/go-dist-scheduler/internal/domain"
	"github.com/yourname/go-dist-scheduler/internal/infrastructure/memory"
	"github.com/yourname/go-dist-scheduler/internal/usecase"
)

func main() {
	jobRepo := memory.NewInMemoryJobRepository()
	taskRepo := memory.NewInMemoryTaskRepository()

	scheduler := usecase.NewScheduler(taskRepo, jobRepo)
	executor := usecase.NewExecutor(jobRepo)

	// Register a sample task
	task := &domain.Task{
		ID:              uuid.New().String(),
		CronExpression:  "* * * * *", // Every minute
		Status:          domain.TaskStatusActive,
		Version:         1,
		LastCheckedAt:   time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := taskRepo.Save(context.Background(), task); err != nil {
		log.Fatalf("failed to save task: %v", err)
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for now := range ticker.C {
		if err := scheduler.CheckAndEnqueue(context.Background(), now); err != nil {
			log.Printf("error in scheduler: %v", err)
		}

		if err := executor.RunPendingJob(context.Background()); err != nil {
			log.Printf("error in executor: %v", err)
		}
	}
}
