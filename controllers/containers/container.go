package containers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/docker/docker/api/types/container"

	"github.com/austiecodes/dws/resources"

	"github.com/gin-gonic/gin"
)

func GetRuningContainers(c *gin.Context) {
	containers, err := resources.DockerClient.ContainerList(context.Background(), container.ListOptions{All: false})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, container := range containers {
		fmt.Printf("Container ID: %s, Image: %s, Names: %v\n", container.ID[:10], container.Image, container.Names)
	}
	c.JSON(http.StatusOK, containers)
}
