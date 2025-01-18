package types

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type ContainerIDrequest struct {
	ContainerID string `json:"containerID"`
}

type CreateContainerOptions struct {
	ContainerConfig  *container.Config         `json:"containerConfig"`
	HostConfig       *container.HostConfig     `json:"HostConfig"`
	NetworkingConfig *network.NetworkingConfig `json:"networkingConfig"`
	Platform         *ocispec.Platform         `json:"platform"`
	ContainerName    string                    `json:"containerName"`
}
