package container

import (
	"fmt"

	"github.com/austiecodes/dws/lib/resources"
	"github.com/austiecodes/dws/models/types"
	dtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
)

func ListContainers(c *gin.Context, uuid string, options container.ListOptions) ([]dtypes.Container, error) {
	containerList, err := resources.DockerClient.ContainerList(c, options)
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("failed to list container: %v", err))
		return nil, err
	}
	storedContainers, err := FetchContainerByUUID(c, uuid)
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("failed to fetch container: %v", err))
		return nil, err
	}
	// check stored containers and container ids are equal
	for _, container := range containerList {
		for _, storedContainer := range storedContainers {
			if container.ID == storedContainer.ContainerID {
				container.Names = append(container.Names, storedContainer.Name)
			}
		}
	}
	return containerList, nil
}

func StartContainerService(c *gin.Context, uuid, containerID string) error {
	err := checkContainerID(c, uuid, containerID)
	if err != nil {
		return err
	}
	err = resources.DockerClient.ContainerStart(c, containerID, container.StartOptions{})
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("failed to start container: %v", err))
		return err
	}
	return nil
}

func StopContainerService(c *gin.Context, uuid, containerID string) error {
	err := checkContainerID(c, uuid, containerID)
	if err != nil {
		return err
	}
	timeout := 10
	err = resources.DockerClient.ContainerStop(c, containerID, container.StopOptions{Timeout: &timeout})
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("failed to stop container: %v", err))
		return err
	}
	return nil
}

func RemoveContainerService(c *gin.Context, uuid, containerID string) error {
	err := checkContainerID(c, uuid, containerID)
	if err != nil {
		return err
	}
	err = RemoveContainerByID(c, containerID)
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("failed to remove container: %v", err))
		return err
	}
	return nil
}

func CreateContainerService(c *gin.Context, uuid string, config *types.CreateContainerOptions) (*container.CreateResponse, error) {
	resp, err := CreateContainer(c, uuid, config)
	if err != nil {
		resources.Logger.Error(fmt.Sprintf("failed to create container: %v", err))
		return nil, err
	}
	return resp, nil
}
