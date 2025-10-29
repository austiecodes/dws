package types

import (
	"time"

	libdb "github.com/austiecodes/dws/internal/lib/db"
)

type ContainerResponse struct {
	ID          uint   `json:"id"`
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	HostSSHPort int    `json:"host_ssh_port"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

func NewContainerResponse(container *libdb.Container) ContainerResponse {
	return ContainerResponse{
		ID:          container.ID,
		UUID:        container.UUID,
		Name:        container.Name,
		Image:       container.Image,
		HostSSHPort: container.HostSSHPort,
		Status:      container.Status,
		CreatedAt:   container.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func NewContainerListResponse(containers []libdb.Container) []ContainerResponse {
	items := make([]ContainerResponse, 0, len(containers))
	for i := range containers {
		items = append(items, NewContainerResponse(&containers[i]))
	}
	return items
}
