package container

import (
	"fmt"
	"net/http"

	"github.com/docker/docker/api/types/container"

	"github.com/austiecodes/dws/models/types"

	"github.com/gin-gonic/gin"
)

func ListContainerController(c *gin.Context) {
	var req types.ContainerIDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}
	containers, err := ListContainers(c, req.UUID, container.ListOptions{All: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, containers)
}

func ListRunningContainerController(c *gin.Context) {
	var req types.ContainerIDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}
	containers, err := ListContainers(c, req.UUID, container.ListOptions{All: false})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, containers)
}

func StartContainerController(c *gin.Context) {
	var req types.ContainerIDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if err := StartContainerService(c, req.UUID, req.ContainerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to start container: %v", err)})
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Container %s started successfully", req.ContainerID)})
}

func StopContainerController(c *gin.Context) {
	var req types.ContainerIDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if err := StopContainerService(c, req.UnixName, req.ContainerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to stop container: %v", err)})
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Container %s stopped successfully", req.ContainerID)})
}

func CreateContaineController(c *gin.Context) {
	var req types.CreateContainerReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	resp, err := CreateContainerService(c, req.UUID, &req.Options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create container: %v", err)})
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Container %s created successfully, id:%v ", req.Options.ContainerName, resp.ID)})
}

func RemoveContainerController(c *gin.Context) {
	var req types.ContainerIDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if err := RemoveContainerService(c, req.UUID, req.ContainerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to remove container: %v", err)})
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Container %s removed successfully", req.ContainerID)})
}
