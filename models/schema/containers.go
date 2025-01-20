package schema

import "time"

type Container struct {
	ID          uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UUID        string    `gorm:"column:uuid" json:"uuid"`
	ContainerID string    `gorm:"column:container_id" json:"container_id"`
	Name        string    `gorm:"column:name" json:"name"`
	Image       string    `gorm:"column:image" json:"image"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}
