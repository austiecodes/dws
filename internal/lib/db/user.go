package db

import "time"

// User represents a platform user that can log into the system.
type User struct {
	ID           uint       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Email        string     `gorm:"column:email;uniqueIndex:users_email_key;not null" json:"email"`
	DisplayName  string     `gorm:"column:display_name;not null" json:"display_name"`
	PasswordHash string     `gorm:"column:password_hash;not null" json:"-"`
	IsAdmin      bool       `gorm:"column:is_admin;not null;default:false" json:"is_admin"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at" json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
