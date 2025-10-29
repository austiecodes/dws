package db

import "time"

type Container struct {
	ID          uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UUID        string    `gorm:"column:uuid;uniqueIndex;not null" json:"uuid"`
	ContainerID string    `gorm:"column:container_id;uniqueIndex;not null" json:"container_id"`
	Name        string    `gorm:"column:name;not null" json:"name"`
	Image       string    `gorm:"column:image;not null" json:"image"`
	UserID      uint      `gorm:"column:user_id;index;not null" json:"user_id"`
	HostSSHPort int       `gorm:"column:host_ssh_port;uniqueIndex;not null" json:"host_ssh_port"`
	Status      string    `gorm:"column:status;not null" json:"status"`
	IsDeleted   bool      `gorm:"column:is_deleted;default:false;not null" json:"is_deleted"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Container) TableName() string {
	return "containers"
}
