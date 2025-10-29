package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	libconfig "github.com/austiecodes/dws/internal/lib/config"
	libdb "github.com/austiecodes/dws/internal/lib/db"
	libdocker "github.com/austiecodes/dws/internal/lib/docker"
	"github.com/austiecodes/dws/internal/platform/repository"
)

var (
	ErrServiceNotInitialised = errors.New("container service not initialised")
	ErrNoAvailablePorts      = errors.New("no available ssh ports")
)

type ContainerService struct {
	cfg libconfig.DockerConfig
}

var Containers *ContainerService

func InitContainerService(cfg libconfig.DockerConfig) {
	Containers = &ContainerService{cfg: cfg}
}

func (s *ContainerService) List(ctx context.Context, userID uint) ([]libdb.Container, error) {
	if s == nil {
		return nil, ErrServiceNotInitialised
	}
	containers, err := repository.Containers.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 同步每个容器的状态
	for i := range containers {
		_ = s.syncContainerStatus(ctx, &containers[i])
	}

	return containers, nil
}

func (s *ContainerService) syncContainerStatus(ctx context.Context, container *libdb.Container) error {
	manager := libdocker.MustInstance()
	status, err := manager.InspectContainer(ctx, container.ContainerID)
	if err != nil {
		return err
	}

	// 如果状态不一致，更新数据库
	if container.Status != status.Status {
		if err := repository.Containers.UpdateStatus(ctx, container.ContainerID, status.Status); err != nil {
			return err
		}
		container.Status = status.Status
	}

	return nil
}

func (s *ContainerService) SyncAllContainers(ctx context.Context) error {
	if s == nil {
		return ErrServiceNotInitialised
	}

	containers, err := repository.Containers.ListAll(ctx)
	if err != nil {
		return err
	}

	manager := libdocker.MustInstance()
	for i := range containers {
		status, err := manager.InspectContainer(ctx, containers[i].ContainerID)
		if err != nil {
			continue // 跳过错误的容器
		}

		if containers[i].Status != status.Status {
			_ = repository.Containers.UpdateStatus(ctx, containers[i].ContainerID, status.Status)
		}
	}

	return nil
}

type CreateContainerOptions struct {
	Image    string
	Password string // Optional SSH password, default: "dws"
}

func (s *ContainerService) Create(ctx context.Context, userID uint, image string) (*libdb.Container, error) {
	return s.CreateWithOptions(ctx, userID, CreateContainerOptions{
		Image:    image,
		Password: "", // Use default password
	})
}

func (s *ContainerService) CreateWithOptions(ctx context.Context, userID uint, opts CreateContainerOptions) (*libdb.Container, error) {
	if s == nil {
		return nil, ErrServiceNotInitialised
	}
	image := strings.TrimSpace(opts.Image)
	if image == "" {
		return nil, errors.New("image is required")
	}

	port, err := s.allocatePort(ctx)
	if err != nil {
		return nil, err
	}

	manager := libdocker.MustInstance()
	name := fmt.Sprintf("dws-u%d-%d", userID, time.Now().Unix())

	// 如果未指定密码，使用默认值 "dws"
	password := opts.Password
	if password == "" {
		password = "dws"
	}

	result, err := manager.CreateSSHContainer(ctx, libdocker.CreateContainerOptions{
		Name:     name,
		Image:    image,
		HostPort: port,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	record := &libdb.Container{
		UUID:        uuid.NewString(),
		ContainerID: result.ID,
		Name:        name,
		Image:       image,
		UserID:      userID,
		HostSSHPort: port,
		Status:      "running",
	}

	if err := repository.Containers.Create(ctx, record); err != nil {
		_ = manager.RemoveContainer(ctx, result.ID)
		return nil, err
	}

	return record, nil
}

func (s *ContainerService) allocatePort(ctx context.Context) (int, error) {
	used, err := repository.Containers.ListHostPorts(ctx, s.cfg.SSHPortRangeStart, s.cfg.SSHPortRangeEnd)
	if err != nil {
		return 0, fmt.Errorf("list used ports: %w", err)
	}

	for port := s.cfg.SSHPortRangeStart; port <= s.cfg.SSHPortRangeEnd; port++ {
		if _, ok := used[port]; ok {
			continue
		}
		if !libdocker.IsPortFree(port) {
			continue
		}
		return port, nil
	}
	return 0, ErrNoAvailablePorts
}

func (s *ContainerService) AllowedImages(ctx context.Context) ([]string, error) {
	if s == nil {
		return nil, ErrServiceNotInitialised
	}
	manager := libdocker.MustInstance()
	return manager.ListImages(ctx)
}

func (s *ContainerService) Stop(ctx context.Context, userID uint, uuid string) error {
	if s == nil {
		return ErrServiceNotInitialised
	}

	// 获取容器并验证所有权
	container, err := repository.Containers.GetByUUID(ctx, uuid)
	if err != nil {
		return fmt.Errorf("get container: %w", err)
	}
	if container.UserID != userID {
		return errors.New("permission denied")
	}

	// 停止容器
	manager := libdocker.MustInstance()
	if err := manager.StopContainer(ctx, container.ContainerID); err != nil {
		return fmt.Errorf("stop container: %w", err)
	}

	// 更新状态
	if err := repository.Containers.UpdateStatus(ctx, container.ContainerID, "stopped"); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	return nil
}

func (s *ContainerService) Start(ctx context.Context, userID uint, uuid string) error {
	if s == nil {
		return ErrServiceNotInitialised
	}

	// 获取容器并验证所有权
	container, err := repository.Containers.GetByUUID(ctx, uuid)
	if err != nil {
		return fmt.Errorf("get container: %w", err)
	}
	if container.UserID != userID {
		return errors.New("permission denied")
	}

	// 启动容器
	manager := libdocker.MustInstance()
	if err := manager.StartContainer(ctx, container.ContainerID); err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	// 更新状态
	if err := repository.Containers.UpdateStatus(ctx, container.ContainerID, "running"); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	return nil
}

func (s *ContainerService) Delete(ctx context.Context, userID uint, uuid string) error {
	if s == nil {
		return ErrServiceNotInitialised
	}

	// 获取容器并验证所有权
	container, err := repository.Containers.GetByUUID(ctx, uuid)
	if err != nil {
		return fmt.Errorf("get container: %w", err)
	}
	if container.UserID != userID {
		return errors.New("permission denied")
	}

	// 开启事务
	conn := libdb.MustInstance()
	tx := conn.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin transaction: %w", tx.Error)
	}

	// 删除 Docker 容器
	manager := libdocker.MustInstance()
	if err := manager.RemoveContainer(ctx, container.ContainerID); err != nil {
		tx.Rollback()
		return fmt.Errorf("remove container: %w", err)
	}

	// 软删除数据库记录
	if err := tx.Model(&libdb.Container{}).
		Where("uuid = ?", uuid).
		Update("is_deleted", true).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("soft delete from db: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
