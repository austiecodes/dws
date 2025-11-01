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

func (r *UserRepository) List(ctx context.Context) ([]libdb.User, error) {
	conn := libdb.MustInstance()
	var users []libdb.User
	if err := conn.WithContext(ctx).Order("created_at ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	conn := libdb.MustInstance()
	var count int64
	if err := conn.WithContext(ctx).Model(&libdb.User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *UserRepository) CountAdmins(ctx context.Context) (int64, error) {
	conn := libdb.MustInstance()
	var count int64
	if err := conn.WithContext(ctx).Model(&libdb.User{}).
		Where("is_admin = ?", true).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *UserRepository) SetAdmin(ctx context.Context, id uint, isAdmin bool) error {
	conn := libdb.MustInstance()
	result := conn.WithContext(ctx).
		Model(&libdb.User{}).
		Where("id = ?", id).
		Update("is_admin", isAdmin)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
