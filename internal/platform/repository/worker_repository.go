package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	libdb "github.com/austiecodes/dws/internal/lib/db"
)

func WorkerCreate(ctx context.Context, db *gorm.DB, worker *libdb.Worker) error {
	return dbWithContext(ctx, db).Create(worker).Error
}

func WorkerGetByID(ctx context.Context, db *gorm.DB, id string) (*libdb.Worker, error) {
	var worker libdb.Worker
	if err := dbWithContext(ctx, db).Where("id = ?", id).First(&worker).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &worker, nil
}

func WorkerList(ctx context.Context, db *gorm.DB) ([]libdb.Worker, error) {
	var workers []libdb.Worker
	if err := dbWithContext(ctx, db).Order("created_at DESC").Find(&workers).Error; err != nil {
		return nil, err
	}
	return workers, nil
}

func WorkerUpdate(ctx context.Context, db *gorm.DB, id string, updates map[string]interface{}) error {
	if updates == nil {
		return nil
	}
	if _, ok := updates["updated_at"]; !ok {
		updates["updated_at"] = time.Now()
	}
	return dbWithContext(ctx, db).
		Model(&libdb.Worker{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func WorkerUpdateHeartbeat(ctx context.Context, db *gorm.DB, id string, status libdb.WorkerStatus) error {
	return WorkerUpdate(ctx, db, id, map[string]interface{}{
		"last_heartbeat": time.Now(),
		"status":         status,
	})
}

func WorkerDelete(ctx context.Context, db *gorm.DB, id string) error {
	return dbWithContext(ctx, db).
		Where("id = ?", id).
		Delete(&libdb.Worker{}).Error
}

func WorkerCountContainers(ctx context.Context, db *gorm.DB, workerID string) (int64, error) {
	var count int64
	err := dbWithContext(ctx, db).
		Model(&libdb.Container{}).
		Where("worker_id = ? AND is_deleted = false", workerID).
		Count(&count).Error
	return count, err
}
