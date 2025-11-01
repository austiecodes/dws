package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	libdb "github.com/austiecodes/dws/internal/lib/db"
)

func UserFindByEmail(ctx context.Context, db *gorm.DB, email string) (*libdb.User, error) {
	var user libdb.User
	result := dbWithContext(ctx, db).Where("email = ?", email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func UserFindByID(ctx context.Context, db *gorm.DB, id uint) (*libdb.User, error) {
	var user libdb.User
	result := dbWithContext(ctx, db).First(&user, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func UserUpdateLastLogin(ctx context.Context, db *gorm.DB, id uint, ts time.Time) error {
	return dbWithContext(ctx, db).
		Model(&libdb.User{}).
		Where("id = ?", id).
		Update("last_login_at", ts).Error
}

func UserCreate(ctx context.Context, db *gorm.DB, user *libdb.User) error {
	return dbWithContext(ctx, db).Create(user).Error
}

func UserList(ctx context.Context, db *gorm.DB) ([]libdb.User, error) {
	var users []libdb.User
	if err := dbWithContext(ctx, db).Order("created_at ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func UserCount(ctx context.Context, db *gorm.DB) (int64, error) {
	var count int64
	if err := dbWithContext(ctx, db).Model(&libdb.User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func UserCountAdmins(ctx context.Context, db *gorm.DB) (int64, error) {
	var count int64
	if err := dbWithContext(ctx, db).
		Model(&libdb.User{}).
		Where("is_admin = ?", true).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func UserSetAdmin(ctx context.Context, db *gorm.DB, id uint, isAdmin bool) error {
	result := dbWithContext(ctx, db).
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
