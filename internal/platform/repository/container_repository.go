package repository

import (
	"context"

	"gorm.io/gorm"

	libdb "github.com/austiecodes/dws/internal/lib/db"
)

func ContainerCreate(ctx context.Context, db *gorm.DB, container *libdb.Container) error {
	return dbWithContext(ctx, db).Create(container).Error
}

func ContainerListByUser(ctx context.Context, db *gorm.DB, userID uint) ([]libdb.Container, error) {
	var containers []libdb.Container
	err := dbWithContext(ctx, db).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Order("created_at DESC").
		Find(&containers).Error
	return containers, err
}

func ContainerListHostPorts(ctx context.Context, db *gorm.DB, start, end int) (map[int]struct{}, error) {
	var ports []int
	if err := dbWithContext(ctx, db).
		Model(&libdb.Container{}).
		Where("host_ssh_port BETWEEN ? AND ?", start, end).
		Pluck("host_ssh_port", &ports).Error; err != nil {
		return nil, err
	}
	return asSet(ports), nil
}

func ContainerListHostPortsForWorker(ctx context.Context, db *gorm.DB, workerID string, start, end int) (map[int]struct{}, error) {
	query := dbWithContext(ctx, db).
		Model(&libdb.Container{}).
		Where("host_ssh_port BETWEEN ? AND ?", start, end)
	if workerID != "" {
		query = query.Where("worker_id = ?", workerID)
	}
	var ports []int
	if err := query.Pluck("host_ssh_port", &ports).Error; err != nil {
		return nil, err
	}
	return asSet(ports), nil
}

func ContainerUpdateStatus(ctx context.Context, db *gorm.DB, containerID string, status string) error {
	return dbWithContext(ctx, db).
		Model(&libdb.Container{}).
		Where("container_id = ?", containerID).
		Update("status", status).Error
}

func ContainerListAll(ctx context.Context, db *gorm.DB) ([]libdb.Container, error) {
	var containers []libdb.Container
	err := dbWithContext(ctx, db).Find(&containers).Error
	return containers, err
}

func ContainerGetByID(ctx context.Context, db *gorm.DB, id uint) (*libdb.Container, error) {
	var container libdb.Container
	if err := dbWithContext(ctx, db).Where("id = ?", id).First(&container).Error; err != nil {
		return nil, err
	}
	return &container, nil
}

func ContainerGetByUUID(ctx context.Context, db *gorm.DB, uuid string) (*libdb.Container, error) {
	var container libdb.Container
	if err := dbWithContext(ctx, db).Where("uuid = ?", uuid).First(&container).Error; err != nil {
		return nil, err
	}
	return &container, nil
}

func ContainerSoftDelete(ctx context.Context, db *gorm.DB, uuid string) error {
	return dbWithContext(ctx, db).
		Model(&libdb.Container{}).
		Where("uuid = ?", uuid).
		Update("is_deleted", true).Error
}

func ContainerHardDelete(ctx context.Context, db *gorm.DB, uuid string) error {
	return dbWithContext(ctx, db).Where("uuid = ?", uuid).Delete(&libdb.Container{}).Error
}

func asSet(values []int) map[int]struct{} {
	used := make(map[int]struct{}, len(values))
	for _, v := range values {
		used[v] = struct{}{}
	}
	return used
}
