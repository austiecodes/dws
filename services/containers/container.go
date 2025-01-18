package services

import (
	"fmt"

	"github.com/austiecodes/dws/models/types"
	dtypes "github.com/docker/docker/api/types"

	"github.com/austiecodes/dws/libs/resources"
	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
)

func ListContainers(c *gin.Context, options container.ListOptions) ([]dtypes.Container, error) {
	containers, err := resources.DockerClient.ContainerList(c, options)
	return containers, err
}

func StartContainerService(c *gin.Context, containerID string) error {
	err := resources.DockerClient.ContainerStart(c, containerID, container.StartOptions{})
	if err != nil {
		c.Error(fmt.Errorf("failed to start container: %v", err))
	}
	return err
}

func StopContainerService(c *gin.Context, containerID string) error {
	timeout := 10
	err := resources.DockerClient.ContainerStop(c, containerID, container.StopOptions{Timeout: &timeout})
	if err != nil {
		c.Error(fmt.Errorf("failed to stop container: %v", err))
		return err
	}
	return nil
}

func RemoveContainerService(c *gin.Context, containerID string) error {
	err := resources.DockerClient.ContainerRemove(c, containerID, container.RemoveOptions{Force: true})
	if err != nil {
		c.Error(fmt.Errorf("failed to remove container: %v", err))
		return err
	}
	return nil
}

func CreateContainerService(c *gin.Context, config *types.CreateContainerOptions) (*container.CreateResponse, error) {
	resp, err := resources.DockerClient.ContainerCreate(c, config.ContainerConfig, config.HostConfig, config.NetworkingConfig, config.Platform, config.ContainerName)
	if err != nil {
		c.Error(fmt.Errorf("failed to create container: %v", err))
		return nil, err
	}
	return &resp, err
}
