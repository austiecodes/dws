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
func (s *TaskService) Create(ctx context.Context, userID uint, containerID uint, command string, expectedDuration, priority int) (*libdb.Task, error) {
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
