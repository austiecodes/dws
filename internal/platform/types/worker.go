package types

import (
	"time"

	libdb "github.com/austiecodes/dws/internal/lib/db"
)

// CreateWorkerRequest represents the request to create a new worker.
type CreateWorkerRequest struct {
	ID       string                 `json:"id" binding:"required,min=1,max=64"`
	Name     string                 `json:"name" binding:"required,min=1,max=255"`
	Address  string                 `json:"address" binding:"required,min=1,max=255"`
	Metadata map[string]interface{} `json:"metadata"`
}

// UpdateWorkerRequest represents the request to update a worker.
type UpdateWorkerRequest struct {
	Name     *string                 `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Address  *string                 `json:"address,omitempty" binding:"omitempty,min=1,max=255"`
	Status   *string                 `json:"status,omitempty" binding:"omitempty,oneof=online offline maintenance"`
	Metadata *map[string]interface{} `json:"metadata,omitempty"`
}

// WorkerResponse represents a worker in API responses.
type WorkerResponse struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Address       string                 `json:"address"`
	Status        string                 `json:"status"`
	LastHeartbeat *time.Time             `json:"last_heartbeat,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	ContainerCount int64                 `json:"container_count,omitempty"`
}

// NewWorkerResponse creates a WorkerResponse from a Worker model.
func NewWorkerResponse(worker *libdb.Worker) WorkerResponse {
	metadata := make(map[string]interface{})
	if len(worker.Metadata) > 0 {
		_ = worker.Metadata.UnmarshalJSON(worker.Metadata)
	}

	return WorkerResponse{
		ID:            worker.ID,
		Name:          worker.Name,
		Address:       worker.Address,
		Status:        string(worker.Status),
		LastHeartbeat: worker.LastHeartbeat,
		Metadata:      metadata,
		CreatedAt:     worker.CreatedAt,
		UpdatedAt:     worker.UpdatedAt,
	}
}

// NewWorkerListResponse creates a list of WorkerResponse from Worker models.
func NewWorkerListResponse(workers []libdb.Worker) []WorkerResponse {
	responses := make([]WorkerResponse, len(workers))
	for i, worker := range workers {
		responses[i] = NewWorkerResponse(&worker)
	}
	return responses
}

