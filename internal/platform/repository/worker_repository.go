package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	libdb "github.com/austiecodes/dws/internal/lib/db"
)

type WorkerRepository struct{}

var Workers = &WorkerRepository{}

// Create creates a new worker.
func (r *WorkerRepository) Create(worker *libdb.Worker) error {
	conn := libdb.MustInstance()
	return conn.Create(worker).Error
}

// GetByID retrieves a worker by ID.
func (r *WorkerRepository) GetByID(id string) (*libdb.Worker, error) {
	conn := libdb.MustInstance()
	var worker libdb.Worker
	if err := conn.Where("id = ?", id).First(&worker).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &worker, nil
}

// List returns all workers.
func (r *WorkerRepository) List() ([]libdb.Worker, error) {
	conn := libdb.MustInstance()
	var workers []libdb.Worker
	if err := conn.Order("created_at DESC").Find(&workers).Error; err != nil {
		return nil, err
	}
	return workers, nil
}

// Update updates a worker's fields.
func (r *WorkerRepository) Update(id string, updates map[string]interface{}) error {
	conn := libdb.MustInstance()
	updates["updated_at"] = time.Now()
	return conn.Model(&libdb.Worker{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateHeartbeat updates the last heartbeat timestamp.
func (r *WorkerRepository) UpdateHeartbeat(id string) error {
	conn := libdb.MustInstance()
	return conn.Model(&libdb.Worker{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_heartbeat": time.Now(),
			"status":         libdb.WorkerStatusOnline,
		}).Error
}

// Delete deletes a worker by ID.
func (r *WorkerRepository) Delete(id string) error {
	conn := libdb.MustInstance()
	return conn.Where("id = ?", id).Delete(&libdb.Worker{}).Error
}

// CountContainers returns the number of containers on a worker.
func (r *WorkerRepository) CountContainers(workerID string) (int64, error) {
	conn := libdb.MustInstance()
	var count int64
	err := conn.Model(&libdb.Container{}).
		Where("worker_id = ? AND is_deleted = false", workerID).
		Count(&count).Error
	return count, err
}

