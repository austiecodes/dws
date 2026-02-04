package handlers

import (
	"encoding/json"
	"time"

	libdb "github.com/austiecodes/dws/internal/lib/db"
)

type containerJSON struct {
	ID          uint   `json:"id"`
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	HostSSHPort int    `json:"host_ssh_port"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

func newContainerJSON(container *libdb.Container) containerJSON {
	return containerJSON{
		ID:          container.ID,
		UUID:        container.UUID,
		Name:        container.Name,
		Image:       container.Image,
		HostSSHPort: container.HostSSHPort,
		Status:      container.Status,
		CreatedAt:   container.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func newContainerListJSON(containers []libdb.Container) []containerJSON {
	items := make([]containerJSON, 0, len(containers))
	for i := range containers {
		items = append(items, newContainerJSON(&containers[i]))
	}
	return items
}

type userJSON struct {
	ID          uint   `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	IsAdmin     bool   `json:"is_admin"`
}

func newUserJSON(user *libdb.User) userJSON {
	return userJSON{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		IsAdmin:     user.IsAdmin,
	}
}

type workerJSON struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Address        string                 `json:"address"`
	Status         string                 `json:"status"`
	IsOnline       bool                   `json:"is_online"`
	LastHeartbeat  *time.Time             `json:"last_heartbeat,omitempty"`
	LeaseExpiresAt *time.Time             `json:"lease_expires_at,omitempty"`
	Metadata       map[string]interface{} `json:"metadata"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	ContainerCount int64                  `json:"container_count,omitempty"`
}

func newWorkerJSON(worker *libdb.Worker) workerJSON {
	metadata := make(map[string]interface{})
	if len(worker.Metadata) > 0 {
		_ = json.Unmarshal(worker.Metadata, &metadata)
	}

	isOnline := false
	if worker.LeaseExpiresAt != nil && time.Now().Before(*worker.LeaseExpiresAt) && worker.Status != libdb.WorkerStatusMaintenance {
		isOnline = true
	}

	status := worker.Status
	if !isOnline && status == libdb.WorkerStatusOnline {
		status = libdb.WorkerStatusOffline
	}

	return workerJSON{
		ID:             worker.ID,
		Name:           worker.Name,
		Address:        worker.Address,
		Status:         string(status),
		IsOnline:       isOnline,
		LastHeartbeat:  worker.LastHeartbeat,
		LeaseExpiresAt: worker.LeaseExpiresAt,
		Metadata:       metadata,
		CreatedAt:      worker.CreatedAt,
		UpdatedAt:      worker.UpdatedAt,
	}
}

func newWorkerListJSON(workers []libdb.Worker) []workerJSON {
	items := make([]workerJSON, len(workers))
	for i := range workers {
		items[i] = newWorkerJSON(&workers[i])
	}
	return items
}

type taskJSON struct {
	ID               uint           `json:"id"`
	UserID           uint           `json:"user_id"`
	ContainerID      uint           `json:"container_id"`
	Command          string         `json:"command"`
	Status           string         `json:"status"`
	TaskType         string         `json:"task_type"`
	ExpectedDuration int            `json:"expected_duration"`
	Priority         int            `json:"priority"`
	StartedAt        *string        `json:"started_at,omitempty"`
	CompletedAt      *string        `json:"completed_at,omitempty"`
	Output           string         `json:"output"`
	ExitCode         *int           `json:"exit_code,omitempty"`
	CreatedAt        string         `json:"created_at"`
	Container        *containerJSON `json:"container,omitempty"`
}

func newTaskJSON(task *libdb.Task) taskJSON {
	resp := taskJSON{
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
		container := newContainerJSON(task.Container)
		resp.Container = &container
	}

	return resp
}

func newTaskListJSON(tasks []libdb.Task) []taskJSON {
	items := make([]taskJSON, 0, len(tasks))
	for i := range tasks {
		items = append(items, newTaskJSON(&tasks[i]))
	}
	return items
}
