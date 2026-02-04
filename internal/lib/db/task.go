package db

import (
	"time"

	"gorm.io/datatypes"
)

// TaskStatus defines the lifecycle states of a task.
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusKilled    TaskStatus = "killed"
)

// TaskType defines the resource type required for task execution.
type TaskType string

const (
	TaskTypeCPU TaskType = "cpu" // CPU tasks: max 3 concurrent
	TaskTypeGPU TaskType = "gpu" // GPU tasks: max 1 concurrent (exclusive)
)

// Task represents a user-submitted job to be executed in a container.
type Task struct {
	ID          uint    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID      uint    `gorm:"column:user_id;index;not null" json:"user_id"`
	ContainerID uint    `gorm:"column:container_id;index;not null" json:"container_id"`
	WorkerID    *string `gorm:"column:worker_id;index" json:"worker_id,omitempty"`

	Command          string         `gorm:"column:command;not null" json:"command"`
	TaskType         TaskType       `gorm:"column:task_type;not null;default:'cpu'" json:"task_type"`
	ExpectedDuration int            `gorm:"column:expected_duration;not null" json:"expected_duration"` // seconds
	Priority         int            `gorm:"column:priority;not null;default:0" json:"priority"`
	Status           TaskStatus     `gorm:"column:status;not null;default:'pending'" json:"status"`
	StartedAt        *time.Time     `gorm:"column:started_at" json:"started_at,omitempty"`
	CompletedAt      *time.Time     `gorm:"column:completed_at" json:"completed_at,omitempty"`
	Output           string         `gorm:"column:output;type:text" json:"output"`
	ExitCode         *int           `gorm:"column:exit_code" json:"exit_code,omitempty"`
	Metadata         datatypes.JSON `gorm:"column:metadata;type:jsonb;default:'{}';not null" json:"metadata"`
	CreatedAt        time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Relations (not stored in DB, loaded via GORM preload)
	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Container *Container `gorm:"foreignKey:ContainerID" json:"container,omitempty"`
	Worker    *Worker    `gorm:"foreignKey:WorkerID;references:ID" json:"worker,omitempty"`
}

func (Task) TableName() string {
	return "tasks"
}
