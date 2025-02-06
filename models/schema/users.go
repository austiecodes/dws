package schema

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UUID      string    `gorm:"column:uuid" json:"uuid"`
	UserName  string    `gorm:"column:user_name" json:"user_name"`
	UnixName  string    `gorm:"column:unix_name" json:"unix_name"`
	UserPswd  string    `gorm:"column:user_pswd" json:"user_pswd"`
	Forbidden bool      `gorm:"column:forbidden" json:"forbidden"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}
