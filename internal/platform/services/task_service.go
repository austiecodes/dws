package services

import (
	"context"
	"errors"

	libdb "github.com/austiecodes/dws/internal/lib/db"
	"github.com/austiecodes/dws/internal/platform/repository"
)

var (
	ErrTaskNotFound        = errors.New("task not found")
	ErrContainerNotFound   = errors.New("container not found")
	ErrContainerNotRunning = errors.New("container is not running")
	ErrUnauthorized        = errors.New("unauthorized to access this task")
)

type TaskService struct{}

var Tasks *TaskService

func InitTaskService() {
	Tasks = &TaskService{}
}

// Create validates and creates a new task for the user.
func (s *TaskService) Create(ctx context.Context, userID uint, containerID uint, command string, taskType libdb.TaskType, expectedDuration, priority int) (*libdb.Task, error) {
	// Verify the container exists and belongs to the user
	container, err := repository.Containers.GetByID(ctx, containerID)
	if err != nil {
		return nil, err
	}
	if container == nil {
		return nil, ErrContainerNotFound
	}
	if container.UserID != userID {
		return nil, ErrUnauthorized
	}

	// Check if container is running (optional: you might allow scheduling tasks on stopped containers)
	if container.Status != "running" {
		return nil, ErrContainerNotRunning
	}

	task := &libdb.Task{
		UserID:           userID,
		ContainerID:      containerID,
		Command:          command,
		Status:           libdb.TaskStatusPending,
		TaskType:         taskType,
		ExpectedDuration: expectedDuration,
		Priority:         priority,
	}

	if err := repository.Tasks.Create(task); err != nil {
		return nil, err
	}

	return task, nil
}

// GetByID retrieves a task by ID, ensuring the user owns it.
func (s *TaskService) GetByID(ctx context.Context, taskID, userID uint) (*libdb.Task, error) {
	task, err := repository.Tasks.GetByID(taskID, true)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}
	if task.UserID != userID {
		return nil, ErrUnauthorized
	}
	return task, nil
}

// List returns all tasks for a user.
func (s *TaskService) List(ctx context.Context, userID uint) ([]libdb.Task, error) {
	return repository.Tasks.ListByUser(userID, true)
}

// Cancel allows a user to cancel their own task (only if pending or running).
func (s *TaskService) Cancel(ctx context.Context, taskID, userID uint) error {
	return repository.Tasks.CancelByUser(taskID, userID)
}

// GetQueueStatistics returns system-wide task queue statistics.
func (s *TaskService) GetQueueStatistics(ctx context.Context) (map[string]interface{}, error) {
	runningCounts, err := repository.Tasks.GetRunningCountsByType()
	if err != nil {
		return nil, err
	}

	pendingCounts, err := repository.Tasks.GetPendingCountsByType()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"running": map[string]int{
			"cpu":   runningCounts[libdb.TaskTypeCPU],
			"gpu":   runningCounts[libdb.TaskTypeGPU],
			"total": runningCounts[libdb.TaskTypeCPU] + runningCounts[libdb.TaskTypeGPU],
		},
		"pending": map[string]int{
			"cpu":   pendingCounts[libdb.TaskTypeCPU],
			"gpu":   pendingCounts[libdb.TaskTypeGPU],
			"total": pendingCounts[libdb.TaskTypeCPU] + pendingCounts[libdb.TaskTypeGPU],
		},
		"limits": map[string]int{
			"cpu": 3,
			"gpu": 1,
		},
	}

	return stats, nil
}
