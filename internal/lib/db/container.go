package db

import (
	"time"

	"gorm.io/datatypes"
)

type Container struct {
	ID          uint           `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UUID        string         `gorm:"column:uuid;uniqueIndex;not null" json:"uuid"`
	ContainerID string         `gorm:"column:container_id;uniqueIndex;not null" json:"container_id"`
	Name        string         `gorm:"column:name;not null" json:"name"`
	Image       string         `gorm:"column:image;not null" json:"image"`
	UserID      uint           `gorm:"column:user_id;index;not null" json:"user_id"`
	WorkerID    *string        `gorm:"column:worker_id;index" json:"worker_id,omitempty"` // NULL = local, otherwise remote worker
	HostSSHPort int            `gorm:"column:host_ssh_port;uniqueIndex;not null" json:"host_ssh_port"`
	Status      string         `gorm:"column:status;not null" json:"status"`
	Config      datatypes.JSON `gorm:"column:config;type:jsonb;default:'{}';not null" json:"config"`
	IsDeleted   bool           `gorm:"column:is_deleted;default:false;not null" json:"is_deleted"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Relations
	Worker *Worker `gorm:"foreignKey:WorkerID;references:ID" json:"worker,omitempty"`
}

func (Container) TableName() string {
	return "containers"
}
