package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/austiecodes/dws/internal/platform/services"
	"github.com/austiecodes/dws/internal/platform/types"
)

func CreateWorker(c *gin.Context) {
	adminUser, ok := requireAdminUser(c)
	if !ok {
		return
	}

	var req types.CreateWorkerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	worker, err := services.WorkersService.Create(
		c.Request.Context(),
		adminUser.ID,
		req.ID,
		req.Name,
		req.Address,
		req.Metadata,
	)
	if err != nil {
		switch err {
		case services.ErrNotAdmin:
			c.JSON(http.StatusForbidden, gin.H{"error": "admin permission required"})
			return
		case services.ErrWorkerIDExists:
			c.JSON(http.StatusConflict, gin.H{"error": "worker ID already exists"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	response := types.NewWorkerResponse(worker)
	c.JSON(http.StatusCreated, gin.H{"worker": response})
}

func ListWorkers(c *gin.Context) {
	if _, ok := requireUser(c); !ok {
		return
	}

	workers, err := services.WorkersService.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses := types.NewWorkerListResponse(workers)

	// Optionally attach container counts
	for i := range responses {
		count, err := services.WorkersService.GetContainerCount(c.Request.Context(), responses[i].ID)
		if err == nil {
			responses[i].ContainerCount = count
		}
	}

	c.JSON(http.StatusOK, gin.H{"workers": responses})
}

func GetWorker(c *gin.Context) {
	if _, ok := requireUser(c); !ok {
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "worker id is required"})
		return
	}

	worker, err := services.WorkersService.GetByID(c.Request.Context(), id)
	if err != nil {
		switch err {
		case services.ErrWorkerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "worker not found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	response := types.NewWorkerResponse(worker)

	// Attach container count
	count, err := services.WorkersService.GetContainerCount(c.Request.Context(), worker.ID)
	if err == nil {
		response.ContainerCount = count
	}

	c.JSON(http.StatusOK, gin.H{"worker": response})
}

func UpdateWorker(c *gin.Context) {
	adminUser, ok := requireAdminUser(c)
	if !ok {
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "worker id is required"})
		return
	}

	var req types.UpdateWorkerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Metadata != nil {
		updates["metadata"] = *req.Metadata
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	if err := services.WorkersService.Update(c.Request.Context(), adminUser.ID, id, updates); err != nil {
		switch err {
		case services.ErrNotAdmin:
			c.JSON(http.StatusForbidden, gin.H{"error": "admin permission required"})
			return
		case services.ErrWorkerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "worker not found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "worker updated"})
}

func DeleteWorker(c *gin.Context) {
	adminUser, ok := requireAdminUser(c)
	if !ok {
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "worker id is required"})
		return
	}

	if err := services.WorkersService.Delete(c.Request.Context(), adminUser.ID, id); err != nil {
		switch err {
		case services.ErrNotAdmin:
			c.JSON(http.StatusForbidden, gin.H{"error": "admin permission required"})
			return
		case services.ErrWorkerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "worker not found"})
			return
		case services.ErrWorkerHasContainers:
			c.JSON(http.StatusConflict, gin.H{"error": "worker has active containers"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "worker deleted"})
}
