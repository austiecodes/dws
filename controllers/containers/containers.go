package containers

import (
	"fmt"
	"net/http"

	"github.com/docker/docker/api/types/container"

	"github.com/austiecodes/dws/models/types"
	services "github.com/austiecodes/dws/services/containers"

	"github.com/gin-gonic/gin"
)

func ListContainers(c *gin.Context) {
	containers, err := services.ListContainers(c, container.ListOptions{All: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, containers)
}

func ListRunningContainers(c *gin.Context) {
	containers, err := services.ListContainers(c, container.ListOptions{All: false})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, containers)
}

func StartContainers(c *gin.Context) {
	var body types.ContainerIDrequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if err := services.StartContainerService(c, body.ContainerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to start container: %v", err)})
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Container %s started successfully", body.ContainerID)})
}

func StopContainers(c *gin.Context) {
	var body types.ContainerIDrequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if err := services.StopContainerService(c, body.ContainerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to stop container: %v", err)})
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Container %s stopped successfully", body.ContainerID)})
}

func RemoveContainers(c *gin.Context) {
	var body types.ContainerIDrequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if err := services.RemoveContainerService(c, body.ContainerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to remove container: %v", err)})
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Container %s removed successfully", body.ContainerID)})
}

func CreateContainer(c *gin.Context) {
	var body types.CreateContainerOptions
	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	resp, err := services.CreateContainerService(c, &body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create container: %v", err)})
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Container %s created successfully, id:%v ", body.ContainerName, resp.ID)})
}
