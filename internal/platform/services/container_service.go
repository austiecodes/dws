package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/datatypes"

	"github.com/austiecodes/dws/internal/lib/apis/codec"
	"github.com/austiecodes/dws/internal/lib/apis/workerpb"
	libconfig "github.com/austiecodes/dws/internal/lib/config"
	libdb "github.com/austiecodes/dws/internal/lib/db"
	libdocker "github.com/austiecodes/dws/internal/lib/docker"
	"github.com/austiecodes/dws/internal/platform/repository"
)

var (
	ErrServiceNotInitialised = errors.New("container service not initialised")
	ErrNoAvailablePorts      = errors.New("no available ssh ports")
	ErrNoOnlineWorker        = errors.New("no online worker available")
)

const (
	workerRPCDialTimeout = 10 * time.Second
)

type ContainerService struct {
	cfg libconfig.DockerConfig
}

var Containers *ContainerService

func InitContainerService(cfg libconfig.DockerConfig) {
	codec.Register()
	Containers = &ContainerService{cfg: cfg}
}

func (s *ContainerService) List(ctx context.Context, userID uint) ([]libdb.Container, error) {
	if s == nil {
		return nil, ErrServiceNotInitialised
	}

	containers, err := repository.ContainerListByUser(ctx, nil, userID)
	if err != nil {
		return nil, err
	}

	for i := range containers {
		_ = s.syncContainerStatus(ctx, &containers[i])
	}

	return containers, nil
}

func (s *ContainerService) syncContainerStatus(ctx context.Context, container *libdb.Container) error {
	if container.WorkerID != nil && *container.WorkerID != "" {
		// Remote containers report status through heartbeats; skip direct docker inspection.
		return nil
	}

	manager, err := libdocker.Instance()
	if err != nil {
		return err
	}
	status, err := manager.InspectContainer(ctx, container.ContainerID)
	if err != nil {
		return err
	}

	if container.Status != status.Status {
		if err := repository.ContainerUpdateStatus(ctx, nil, container.ContainerID, status.Status); err != nil {
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

	containers, err := repository.ContainerListAll(ctx, nil)
	if err != nil {
		return err
	}

	manager, err := libdocker.Instance()
	if err != nil {
		return err
	}
	for i := range containers {
		if containers[i].WorkerID != nil && *containers[i].WorkerID != "" {
			continue
		}
		status, err := manager.InspectContainer(ctx, containers[i].ContainerID)
		if err != nil {
			continue
		}
		if containers[i].Status != status.Status {
			_ = repository.ContainerUpdateStatus(ctx, nil, containers[i].ContainerID, status.Status)
		}
	}

	return nil
}

type CreateContainerOptions struct {
	Image    string
	Password string
}

func (s *ContainerService) Create(ctx context.Context, userID uint, image string) (*libdb.Container, error) {
	return s.CreateWithOptions(ctx, userID, CreateContainerOptions{Image: image})
}

func (s *ContainerService) CreateWithOptions(ctx context.Context, userID uint, opts CreateContainerOptions) (*libdb.Container, error) {
	if s == nil {
		return nil, ErrServiceNotInitialised
	}

	image := strings.TrimSpace(opts.Image)
	if image == "" {
		return nil, errors.New("image is required")
	}

	worker, err := s.selectWorker(ctx)
	if err != nil {
		return nil, err
	}

	port, err := s.allocatePortForWorker(ctx, worker)
	if err != nil {
		return nil, err
	}

	password := opts.Password
	if password == "" {
		password = "dws"
	}

	name := fmt.Sprintf("dws-u%d-%d", userID, time.Now().Unix())

	client, conn, err := s.dialWorker(ctx, worker)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := &workerpb.ContainerSpec{
		Image:         image,
		Password:      password,
		Name:          name,
		UserID:        uint64(userID),
		HostSSHPort:   int32(port),
		ContainerUUID: uuid.NewString(),
	}

	rpcCtx, cancel := context.WithTimeout(ctx, workerRPCDialTimeout)
	defer cancel()

	resp, err := client.CreateContainer(rpcCtx, req)
	if err != nil {
		return nil, err
	}

	configData := map[string]interface{}{
		"ssh_password": password,
		"worker_id":    worker.ID,
		"labels": map[string]string{
			"created_by": "dws-platform",
			"user_id":    fmt.Sprintf("%d", userID),
		},
	}
	configJSON, err := json.Marshal(configData)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}

	workerID := worker.ID
	record := &libdb.Container{
		UUID:        req.ContainerUUID,
		ContainerID: resp.ContainerID,
		Name:        name,
		Image:       image,
		UserID:      userID,
		WorkerID:    &workerID,
		HostSSHPort: int(resp.HostSSHPort),
		Status:      resp.Status,
		Config:      datatypes.JSON(configJSON),
	}

	if err := repository.ContainerCreate(ctx, nil, record); err != nil {
		delCtx, cancel := context.WithTimeout(ctx, workerRPCDialTimeout)
		defer cancel()
		_, _ = client.DeleteContainer(delCtx, &workerpb.ContainerRequest{ContainerID: resp.ContainerID, ContainerUUID: req.ContainerUUID})
		return nil, err
	}

	return record, nil
}

func (s *ContainerService) allocatePortForWorker(ctx context.Context, worker *libdb.Worker) (int, error) {
	workerID := ""
	if worker != nil {
		workerID = worker.ID
	}
	used, err := repository.ContainerListHostPortsForWorker(ctx, nil, workerID, s.cfg.SSHPortRangeStart, s.cfg.SSHPortRangeEnd)
	if err != nil {
		return 0, fmt.Errorf("list used ports: %w", err)
	}

	for port := s.cfg.SSHPortRangeStart; port <= s.cfg.SSHPortRangeEnd; port++ {
		if _, ok := used[port]; ok {
			continue
		}
		if workerID == "" && !libdocker.IsPortFree(port) {
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
	if len(s.cfg.AllowedImages) > 0 {
		return s.cfg.AllowedImages, nil
	}
	manager := libdocker.MustInstance()
	return manager.ListImages(ctx)
}

func (s *ContainerService) Stop(ctx context.Context, userID uint, uuid string) error {
	container, worker, err := s.loadUserContainer(ctx, userID, uuid)
	if err != nil {
		return err
	}

	status := "stopped"
	if worker == nil {
		manager, err := libdocker.Instance()
		if err != nil {
			return err
		}
		if err := manager.StopContainer(ctx, container.ContainerID); err != nil {
			return fmt.Errorf("stop container: %w", err)
		}
	} else {
		client, conn, err := s.dialWorker(ctx, worker)
		if err != nil {
			return err
		}
		defer conn.Close()
		rpcCtx, cancel := context.WithTimeout(ctx, workerRPCDialTimeout)
		defer cancel()
		resp, err := client.StopContainer(rpcCtx, &workerpb.ContainerRequest{ContainerID: container.ContainerID, ContainerUUID: container.UUID})
		if err != nil {
			return err
		}
		if resp != nil && resp.Status != "" {
			status = resp.Status
		}
	}

	return repository.ContainerUpdateStatus(ctx, nil, container.ContainerID, status)
}

func (s *ContainerService) Start(ctx context.Context, userID uint, uuid string) error {
	container, worker, err := s.loadUserContainer(ctx, userID, uuid)
	if err != nil {
		return err
	}

	status := "running"
	if worker == nil {
		manager, err := libdocker.Instance()
		if err != nil {
			return err
		}
		if err := manager.StartContainer(ctx, container.ContainerID); err != nil {
			return fmt.Errorf("start container: %w", err)
		}
	} else {
		client, conn, err := s.dialWorker(ctx, worker)
		if err != nil {
			return err
		}
		defer conn.Close()
		rpcCtx, cancel := context.WithTimeout(ctx, workerRPCDialTimeout)
		defer cancel()
		resp, err := client.StartContainer(rpcCtx, &workerpb.ContainerRequest{ContainerID: container.ContainerID, ContainerUUID: container.UUID})
		if err != nil {
			return err
		}
		if resp != nil && resp.Status != "" {
			status = resp.Status
		}
	}

	return repository.ContainerUpdateStatus(ctx, nil, container.ContainerID, status)
}

func (s *ContainerService) Delete(ctx context.Context, userID uint, uuid string) error {
	container, worker, err := s.loadUserContainer(ctx, userID, uuid)
	if err != nil {
		return err
	}

	conn := libdb.MustInstance()
	tx := conn.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin transaction: %w", tx.Error)
	}

	rollback := func(cause error) error {
		tx.Rollback()
		return cause
	}

	if worker == nil {
		manager, err := libdocker.Instance()
		if err != nil {
			return rollback(err)
		}
		if err := manager.RemoveContainer(ctx, container.ContainerID); err != nil {
			return rollback(fmt.Errorf("remove container: %w", err))
		}
	} else {
		client, conn, err := s.dialWorker(ctx, worker)
		if err != nil {
			return rollback(err)
		}
		defer conn.Close()
		rpcCtx, cancel := context.WithTimeout(ctx, workerRPCDialTimeout)
		defer cancel()
		if _, err := client.DeleteContainer(rpcCtx, &workerpb.ContainerRequest{ContainerID: container.ContainerID, ContainerUUID: container.UUID}); err != nil {
			return rollback(err)
		}
	}

	if err := repository.ContainerSoftDelete(ctx, tx, uuid); err != nil {
		return rollback(fmt.Errorf("soft delete from db: %w", err))
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (s *ContainerService) loadUserContainer(ctx context.Context, userID uint, uuid string) (*libdb.Container, *libdb.Worker, error) {
	if uuid == "" {
		return nil, nil, errors.New("uuid is required")
	}

	container, err := repository.ContainerGetByUUID(ctx, nil, uuid)
	if err != nil {
		return nil, nil, fmt.Errorf("get container: %w", err)
	}

	// Check permission: user must own the container or have manage_any permission
	canManage, err := RBAC.CanManageContainer(ctx, userID, uuid)
	if err != nil {
		return nil, nil, fmt.Errorf("check permission: %w", err)
	}
	if !canManage {
		return nil, nil, errors.New("permission denied")
	}

	var worker *libdb.Worker
	if container.WorkerID != nil && *container.WorkerID != "" {
		worker, err = repository.WorkerGetByID(ctx, nil, *container.WorkerID)
		if err != nil {
			return nil, nil, err
		}
		if worker == nil {
			return nil, nil, fmt.Errorf("worker %s not found", *container.WorkerID)
		}
	}

	return container, worker, nil
}

func (s *ContainerService) selectWorker(ctx context.Context) (*libdb.Worker, error) {
	workers, err := repository.WorkerList(ctx, nil)
	if err != nil {
		return nil, err
	}

	var selected *libdb.Worker
	var minContainers int64 = 1<<63 - 1
	for i := range workers {
		worker := workers[i]
		if worker.Status != libdb.WorkerStatusOnline {
			continue
		}
		count, err := repository.WorkerCountContainers(ctx, nil, worker.ID)
		if err != nil {
			log.Printf("[platform] count containers for worker %s: %v", worker.ID, err)
			continue
		}
		if selected == nil || count < minContainers {
			copy := worker
			selected = &copy
			minContainers = count
		}
	}

	if selected == nil {
		return nil, ErrNoOnlineWorker
	}
	return selected, nil
}

func (s *ContainerService) dialWorker(_ context.Context, worker *libdb.Worker) (workerpb.WorkerControlServiceClient, *grpc.ClientConn, error) {
	if worker == nil || worker.Address == "" {
		return nil, nil, errors.New("worker address not configured")
	}

	conn, err := grpc.NewClient(worker.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.CallContentSubtype(codec.JSONCodecName)))
	if err != nil {
		return nil, nil, err
	}

	return workerpb.NewWorkerControlServiceClient(conn), conn, nil
}
