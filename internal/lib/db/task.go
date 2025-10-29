package db

import "time"

// TaskStatus defines the lifecycle states of a task.
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusKilled    TaskStatus = "killed"
)

// Task represents a user-submitted job to be executed in a container.
type Task struct {
	ID               uint       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID           uint       `gorm:"column:user_id;index;not null" json:"user_id"`
	ContainerID      uint       `gorm:"column:container_id;index;not null" json:"container_id"`
	Command          string     `gorm:"column:command;not null" json:"command"`
	Status           TaskStatus `gorm:"column:status;not null;default:'pending'" json:"status"`
	ExpectedDuration int        `gorm:"column:expected_duration;not null" json:"expected_duration"` // seconds
	Priority         int        `gorm:"column:priority;not null;default:0" json:"priority"`
	StartedAt        *time.Time `gorm:"column:started_at" json:"started_at,omitempty"`
	CompletedAt      *time.Time `gorm:"column:completed_at" json:"completed_at,omitempty"`
	Output           string     `gorm:"column:output;type:text" json:"output"`
	ExitCode         *int       `gorm:"column:exit_code" json:"exit_code,omitempty"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Relations (not stored in DB, loaded via GORM preload)
	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Container *Container `gorm:"foreignKey:ContainerID" json:"container,omitempty"`
}

func (Task) TableName() string {
	return "tasks"
}
