package db

import (
	"time"

	"gorm.io/datatypes"
)

// WorkerStatus defines the operational state of a worker node.
type WorkerStatus string

const (
	WorkerStatusOnline      WorkerStatus = "online"
	WorkerStatusOffline     WorkerStatus = "offline"
	WorkerStatusMaintenance WorkerStatus = "maintenance"
)

// Worker represents a distributed worker node that runs Docker containers.
type Worker struct {
	ID      string       `gorm:"column:id;primaryKey" json:"id"`
	Name    string       `gorm:"column:name;not null" json:"name"`
	Address string       `gorm:"column:address;not null" json:"address"`
	Status  WorkerStatus `gorm:"column:status;not null;default:'offline'" json:"status"`

	// Capacity limits (NULL = unlimited)
	MaxContainers *int `gorm:"column:max_containers" json:"max_containers,omitempty"`
	MaxCPUCores   *int `gorm:"column:max_cpu_cores" json:"max_cpu_cores,omitempty"`
	MaxMemoryGB   *int `gorm:"column:max_memory_gb" json:"max_memory_gb,omitempty"`
	MaxGPUCount   *int `gorm:"column:max_gpu_count" json:"max_gpu_count,omitempty"`

	// Runtime state
	LastHeartbeat *time.Time     `gorm:"column:last_heartbeat" json:"last_heartbeat,omitempty"`
	Metadata      datatypes.JSON `gorm:"column:metadata;type:jsonb;default:'{}';not null" json:"metadata"`

	// Timestamps
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Worker) TableName() string {
	return "workers"
}

