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
	var req types.ContainerIDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}
	containers, err := services.ListContainers(c, req.UUID, container.ListOptions{All: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, containers)
}

func ListRunningContainers(c *gin.Context) {
	var req types.ContainerIDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}
	containers, err := services.ListContainers(c, req.UUID, container.ListOptions{All: false})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, containers)
}

func StartContainers(c *gin.Context) {
	var req types.ContainerIDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if err := services.StartContainerService(c, req.UUID, req.ContainerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to start container: %v", err)})
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Container %s started successfully", req.ContainerID)})
}

func StopContainers(c *gin.Context) {
	var req types.ContainerIDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if err := services.StopContainerService(c, req.UnixName, req.ContainerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to stop container: %v", err)})
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Container %s stopped successfully", req.ContainerID)})
}

func CreateContainer(c *gin.Context) {
	var req types.CreateContainerReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	resp, err := services.CreateContainerService(c, req.UUID, &req.Options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create container: %v", err)})
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Container %s created successfully, id:%v ", req.Options.ContainerName, resp.ID)})
}

func RemoveContainers(c *gin.Context) {
	var req types.ContainerIDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if err := services.RemoveContainerService(c, req.UUID, req.ContainerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to remove container: %v", err)})
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Container %s removed successfully", req.ContainerID)})
}
