package repository

import (
	"context"

	libdb "github.com/austiecodes/dws/internal/lib/db"
)

type ContainerRepository struct{}

var Containers = &ContainerRepository{}

func (r *ContainerRepository) Create(ctx context.Context, container *libdb.Container) error {
	conn := libdb.MustInstance()
	return conn.WithContext(ctx).Create(container).Error
}

func (r *ContainerRepository) ListByUser(ctx context.Context, userID uint) ([]libdb.Container, error) {
	conn := libdb.MustInstance()
	var containers []libdb.Container
	if err := conn.WithContext(ctx).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Order("created_at DESC").
		Find(&containers).Error; err != nil {
		return nil, err
	}
	return containers, nil
}

func (r *ContainerRepository) ListHostPorts(ctx context.Context, start, end int) (map[int]struct{}, error) {
	conn := libdb.MustInstance()
	var ports []int
	if err := conn.WithContext(ctx).
		Model(&libdb.Container{}).
		Where("host_ssh_port BETWEEN ? AND ?", start, end).
		Pluck("host_ssh_port", &ports).Error; err != nil {
		return nil, err
	}
	used := make(map[int]struct{}, len(ports))
	for _, port := range ports {
		used[port] = struct{}{}
	}
	return used, nil
}

func (r *ContainerRepository) ListHostPortsForWorker(ctx context.Context, workerID string, start, end int) (map[int]struct{}, error) {
	conn := libdb.MustInstance()
	var ports []int
	query := conn.WithContext(ctx).
		Model(&libdb.Container{}).
		Where("host_ssh_port BETWEEN ? AND ?", start, end)
	if workerID != "" {
		query = query.Where("worker_id = ?", workerID)
	}
	if err := query.Pluck("host_ssh_port", &ports).Error; err != nil {
		return nil, err
	}
	used := make(map[int]struct{}, len(ports))
	for _, port := range ports {
		used[port] = struct{}{}
	}
	return used, nil
}

func (r *ContainerRepository) UpdateStatus(ctx context.Context, containerID string, status string) error {
	conn := libdb.MustInstance()
	return conn.WithContext(ctx).
		Model(&libdb.Container{}).
		Where("container_id = ?", containerID).
		Update("status", status).Error
}

func (r *ContainerRepository) ListAll(ctx context.Context) ([]libdb.Container, error) {
	conn := libdb.MustInstance()
	var containers []libdb.Container
	if err := conn.WithContext(ctx).Find(&containers).Error; err != nil {
		return nil, err
	}
	return containers, nil
}

func (r *ContainerRepository) GetByID(ctx context.Context, id uint) (*libdb.Container, error) {
	conn := libdb.MustInstance()
	var container libdb.Container
	if err := conn.WithContext(ctx).Where("id = ?", id).First(&container).Error; err != nil {
		return nil, err
	}
	return &container, nil
}

func (r *ContainerRepository) GetByUUID(ctx context.Context, uuid string) (*libdb.Container, error) {
	conn := libdb.MustInstance()
	var container libdb.Container
	if err := conn.WithContext(ctx).Where("uuid = ?", uuid).First(&container).Error; err != nil {
		return nil, err
	}
	return &container, nil
}

func (r *ContainerRepository) SoftDelete(ctx context.Context, uuid string) error {
	conn := libdb.MustInstance()
	return conn.WithContext(ctx).
		Model(&libdb.Container{}).
		Where("uuid = ?", uuid).
		Update("is_deleted", true).Error
}

func (r *ContainerRepository) HardDelete(ctx context.Context, uuid string) error {
	conn := libdb.MustInstance()
	return conn.WithContext(ctx).Where("uuid = ?", uuid).Delete(&libdb.Container{}).Error
}
