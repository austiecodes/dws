package types

import (
	"time"

	libdb "github.com/austiecodes/dws/internal/lib/db"
)

type CreateTaskRequest struct {
	ContainerID      uint   `json:"container_id" binding:"required"`
	Command          string `json:"command" binding:"required"`
	TaskType         string `json:"task_type" binding:"required,oneof=cpu gpu"`
	ExpectedDuration int    `json:"expected_duration" binding:"required,min=1"` // seconds
	Priority         int    `json:"priority"`
}

type TaskResponse struct {
	ID               uint               `json:"id"`
	UserID           uint               `json:"user_id"`
	ContainerID      uint               `json:"container_id"`
	Command          string             `json:"command"`
	Status           string             `json:"status"`
	TaskType         string             `json:"task_type"`
	ExpectedDuration int                `json:"expected_duration"`
	Priority         int                `json:"priority"`
	StartedAt        *string            `json:"started_at,omitempty"`
	CompletedAt      *string            `json:"completed_at,omitempty"`
	Output           string             `json:"output"`
	ExitCode         *int               `json:"exit_code,omitempty"`
	CreatedAt        string             `json:"created_at"`
	Container        *ContainerResponse `json:"container,omitempty"`
}

func NewTaskResponse(task *libdb.Task) TaskResponse {
	resp := TaskResponse{
		ID:               task.ID,
		UserID:           task.UserID,
		ContainerID:      task.ContainerID,
		Command:          task.Command,
		Status:           string(task.Status),
		TaskType:         string(task.TaskType),
		ExpectedDuration: task.ExpectedDuration,
		Priority:         task.Priority,
		Output:           task.Output,
		ExitCode:         task.ExitCode,
		CreatedAt:        task.CreatedAt.UTC().Format(time.RFC3339),
	}

	if task.StartedAt != nil {
		s := task.StartedAt.UTC().Format(time.RFC3339)
		resp.StartedAt = &s
	}

	if task.CompletedAt != nil {
		s := task.CompletedAt.UTC().Format(time.RFC3339)
		resp.CompletedAt = &s
	}

	if task.Container != nil {
		c := NewContainerResponse(task.Container)
		resp.Container = &c
	}

	return resp
}

func NewTaskListResponse(tasks []libdb.Task) []TaskResponse {
	items := make([]TaskResponse, 0, len(tasks))
	for i := range tasks {
		items = append(items, NewTaskResponse(&tasks[i]))
	}
	return items
}

// QueueStatistics represents the system-wide task queue status.
type QueueStatistics struct {
	Running struct {
		CPU   int `json:"cpu"`
		GPU   int `json:"gpu"`
		Total int `json:"total"`
	} `json:"running"`
	Pending struct {
		CPU   int `json:"cpu"`
		GPU   int `json:"gpu"`
		Total int `json:"total"`
	} `json:"pending"`
	Limits struct {
		CPU int `json:"cpu"`
		GPU int `json:"gpu"`
	} `json:"limits"`
}
