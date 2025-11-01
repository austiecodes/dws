package services

import (
	"context"
	"encoding/json"
	"errors"

	libdb "github.com/austiecodes/dws/internal/lib/db"
	"github.com/austiecodes/dws/internal/platform/repository"
	"gorm.io/datatypes"
)

var (
	ErrWorkerNotFound      = errors.New("worker not found")
	ErrWorkerIDExists      = errors.New("worker ID already exists")
	ErrWorkerHasContainers = errors.New("worker has active containers")
	ErrNotAdmin            = errors.New("admin permission required")
)

type WorkerService struct{}

var WorkersService *WorkerService

func InitWorkerService() {
	WorkersService = &WorkerService{}
}

// requireAdmin checks if the user is admin, returns error if not.
func (s *WorkerService) requireAdmin(ctx context.Context, userID uint) error {
	user, err := repository.Users.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUnauthorized
	}
	if !user.IsAdmin {
		return ErrNotAdmin
	}
	return nil
}

// Create creates a new worker (admin only).
func (s *WorkerService) Create(ctx context.Context, userID uint, id, name, address string, metadata map[string]interface{}) (*libdb.Worker, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return nil, err
	}

	// Check if ID already exists
	existing, err := repository.Workers.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrWorkerIDExists
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	worker := &libdb.Worker{
		ID:       id,
		Name:     name,
		Address:  address,
		Status:   libdb.WorkerStatusOffline,
		Metadata: datatypes.JSON(metadataJSON),
	}

	if err := repository.Workers.Create(worker); err != nil {
		return nil, err
	}

	return worker, nil
}

// GetByID retrieves a worker by ID.
func (s *WorkerService) GetByID(ctx context.Context, id string) (*libdb.Worker, error) {
	worker, err := repository.Workers.GetByID(id)
	if err != nil {
		return nil, err
	}
	if worker == nil {
		return nil, ErrWorkerNotFound
	}
	return worker, nil
}

// List returns all workers (visible to all users).
func (s *WorkerService) List(ctx context.Context) ([]libdb.Worker, error) {
	return repository.Workers.List()
}

// Update updates a worker (admin only).
func (s *WorkerService) Update(ctx context.Context, userID uint, id string, updates map[string]interface{}) error {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return err
	}

	worker, err := repository.Workers.GetByID(id)
	if err != nil {
		return err
	}
	if worker == nil {
		return ErrWorkerNotFound
	}

	// Convert metadata if provided
	if metadata, ok := updates["metadata"]; ok {
		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		updates["metadata"] = datatypes.JSON(metadataJSON)
	}

	return repository.Workers.Update(id, updates)
}

// Delete deletes a worker (admin only).
func (s *WorkerService) Delete(ctx context.Context, userID uint, id string) error {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return err
	}

	worker, err := repository.Workers.GetByID(id)
	if err != nil {
		return err
	}
	if worker == nil {
		return ErrWorkerNotFound
	}

	// Check if worker has active containers
	count, err := repository.Workers.CountContainers(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrWorkerHasContainers
	}

	return repository.Workers.Delete(id)
}

// GetContainerCount returns the number of containers on a worker.
func (s *WorkerService) GetContainerCount(ctx context.Context, workerID string) (int64, error) {
	return repository.Workers.CountContainers(workerID)
}

