package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	libdb "github.com/austiecodes/dws/internal/lib/db"
	"github.com/austiecodes/dws/internal/platform/services"
)

type createTaskRequest struct {
	ContainerID      uint   `json:"container_id" binding:"required"`
	Command          string `json:"command" binding:"required"`
	TaskType         string `json:"task_type" binding:"required,oneof=cpu gpu"`
	ExpectedDuration int    `json:"expected_duration" binding:"required,min=1"`
	Priority         int    `json:"priority"`
}

func CreateTask(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}

	var req createTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := services.Tasks.Create(
		c.Request.Context(),
		userID,
		req.ContainerID,
		req.Command,
		libdb.TaskType(req.TaskType),
		req.ExpectedDuration,
		req.Priority,
	)
	if err != nil {
		switch err {
		case services.ErrContainerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "container not found"})
			return
		case services.ErrContainerNotRunning:
			c.JSON(http.StatusBadRequest, gin.H{"error": "container is not running"})
			return
		case services.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "container does not belong to you"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	response := newTaskJSON(task)
	c.JSON(http.StatusCreated, gin.H{"task": response})
}

func ListTasks(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}

	tasks, err := services.Tasks.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": newTaskListJSON(tasks)})
}

func GetTask(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	task, err := services.Tasks.GetByID(c.Request.Context(), uint(id), userID)
	if err != nil {
		switch err {
		case services.ErrTaskNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		case services.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"task": newTaskJSON(task)})
}

func CancelTask(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	if err := services.Tasks.Cancel(c.Request.Context(), uint(id), userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "task cancelled"})
}

func GetQueueStatus(c *gin.Context) {
	stats, err := services.Tasks.GetQueueStatistics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
