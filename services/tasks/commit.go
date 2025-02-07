package services

import (
	"fmt"

	dal "github.com/austiecodes/dws/dal/containers"
	"github.com/austiecodes/dws/models/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/gin-gonic/gin"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func CommitTask(c *gin.Context, taskConfig types.TaskConfig, uuid, userName, containerID string) (string, error) {
	// commit container as image
	commitedImageID, err := dal.CommitContainerAsImage(c, uuid, userName, containerID, taskConfig.Comment)
	if err != nil {
		return "", err
	}
	// get container config info: mountes... etc.
	mounts, err := dal.GetContainerMounts(c, uuid, containerID)
	if err != nil {
		return "", err
	}
	var binds []string
	for _, mount := range mounts {
		bind := fmt.Sprintf("%s:%s", mount.Source, mount.Destination)
		binds = append(binds, bind)
	}
	if len(binds) == 0 {
		binds = nil
	}
	// create container
	resp, err := dal.CreateContainer(c, uuid, &types.CreateContainerOptions{
		ContainerConfig: &container.Config{
			Image: commitedImageID,
			Cmd:   []string{"tail", "-f", "/dev/null"},
		},
		HostConfig: &container.HostConfig{
			Binds: binds,
		},
		NetworkingConfig: &network.NetworkingConfig{},
		Platform:         &ocispec.Platform{},
	})
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}
