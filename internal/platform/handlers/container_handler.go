package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/austiecodes/dws/internal/platform/services"
)

type createContainerRequest struct {
	Image    string `json:"image" binding:"required"`
	Password string `json:"password"` // Optional SSH password, default: "dws"
}

func ListContainers(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}

	containers, err := services.Containers.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"containers": newContainerListJSON(containers)})
}

func CreateContainer(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}

	var req createContainerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	container, err := services.Containers.CreateWithOptions(c.Request.Context(), userID, services.CreateContainerOptions{
		Image:    req.Image,
		Password: req.Password,
	})
	if err != nil {
		switch err {
		case services.ErrNoAvailablePorts:
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no available ssh ports"})
			return
		case services.ErrNoOnlineWorker:
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no online worker available"})
			return
		case services.ErrServiceNotInitialised:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "container service not initialised"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"container": newContainerJSON(container)})
}

func ListImages(c *gin.Context) {
	if _, ok := requireUser(c); !ok {
		return
	}
	images, err := services.Containers.AllowedImages(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(images) == 0 {
		c.JSON(http.StatusOK, gin.H{"images": []string{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"images": images})
}

func StopContainer(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}

	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "uuid is required"})
		return
	}

	if err := services.Containers.Stop(c.Request.Context(), userID, uuid); err != nil {
		if err.Error() == "permission denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "container stopped"})
}

func StartContainer(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}

	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "uuid is required"})
		return
	}

	if err := services.Containers.Start(c.Request.Context(), userID, uuid); err != nil {
		if err.Error() == "permission denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "container started"})
}

func DeleteContainer(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}

	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "uuid is required"})
		return
	}

	if err := services.Containers.Delete(c.Request.Context(), userID, uuid); err != nil {
		if err.Error() == "permission denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "container deleted"})
}
