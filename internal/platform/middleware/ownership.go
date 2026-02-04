package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/austiecodes/dws/internal/lib/rbac"
	"github.com/austiecodes/dws/internal/platform/repository"
)

// RequireTaskOwnershipOr checks if user owns the task OR has the specified permission.
// Use for routes like /tasks/:id/cancel where users can cancel their own tasks,
// but admins can cancel any task.
func RequireTaskOwnershipOr(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getCurrentUser(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
			return
		}
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		// Check if user has the "manage any" permission (admin+)
		if rbac.HasPermissionByUserRole(user.Role, "tasks", permission) {
			c.Set(contextUserKey, user)
			c.Set(contextUserRole, user.Role)
			c.Next()
			return
		}

		// Otherwise, check ownership
		taskIDStr := c.Param("id")
		taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
			return
		}

		task, err := repository.TaskGetByID(c.Request.Context(), nil, uint(taskID), false)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to load task"})
			return
		}
		if task == nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}

		if task.UserID != user.ID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "permission denied"})
			return
		}

		c.Set(contextUserKey, user)
		c.Set(contextUserRole, user.Role)
		c.Next()
	}
}

// RequireContainerOwnershipOr checks if user owns the container OR has the specified permission.
// Use for routes like /containers/:uuid/stop where users can stop their own containers,
// but admins can stop any container.
func RequireContainerOwnershipOr(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getCurrentUser(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
			return
		}
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		// Check if user has the "manage any" permission (admin+)
		if rbac.HasPermissionByUserRole(user.Role, "containers", permission) {
			c.Set(contextUserKey, user)
			c.Set(contextUserRole, user.Role)
			c.Next()
			return
		}

		// Otherwise, check ownership
		containerUUID := c.Param("uuid")
		if containerUUID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid container UUID"})
			return
		}

		container, err := repository.ContainerGetByUUID(c.Request.Context(), nil, containerUUID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to load container"})
			return
		}
		if container == nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "container not found"})
			return
		}

		if container.UserID != user.ID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "permission denied"})
			return
		}

		c.Set(contextUserKey, user)
		c.Set(contextUserRole, user.Role)
		c.Next()
	}
}
