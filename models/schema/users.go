package schema

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UUID      uuid.UUID `gorm:"column:uuid" json:"uuid"`
	UnixName  string    `gorm:"column:unix_name" json:"unix_name"`
	UserName  string    `gorm:"column:user_name" json:"user_name"`
	Password  string    `gorm:"column:pswd" json:"pswd"`
	Forbidden bool      `gorm:"column:forbidden" json:"forbidden"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}
