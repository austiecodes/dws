package services

import (
	dal "github.com/austiecodes/dws/dal/containers"
	"github.com/austiecodes/dws/models/types"
	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
)

func CommitTask(c *gin.Context, uuid, userName, containerID, comment string) (string, error) {
	commitedImageID, err := dal.CommitContainerAsImage(c, uuid, userName, containerID, comment)
	if err != nil {
		return "", err
	}
	resp, err := dal.CreateContainer(c, uuid, &types.CreateContainerOptions{
		ContainerConfig: &container.Config{
			Image: commitedImageID,
		},
	})
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}
