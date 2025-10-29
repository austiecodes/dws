package repository

import (
	"context"
	"errors"
	"time"

	libdb "github.com/austiecodes/dws/internal/lib/db"
	"gorm.io/gorm"
)

type UserRepository struct {
}

var Users = &UserRepository{}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*libdb.User, error) {
	var user libdb.User
	conn := libdb.MustInstance()
	result := conn.WithContext(ctx).Where("email = ?", email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uint) (*libdb.User, error) {
	var user libdb.User
	conn := libdb.MustInstance()
	result := conn.WithContext(ctx).First(&user, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uint, ts time.Time) error {
	conn := libdb.MustInstance()
	return conn.WithContext(ctx).
		Model(&libdb.User{}).
		Where("id = ?", id).
		Update("last_login_at", ts).Error
}

func (r *UserRepository) Create(ctx context.Context, user *libdb.User) error {
	conn := libdb.MustInstance()
	return conn.WithContext(ctx).Create(user).Error
}
