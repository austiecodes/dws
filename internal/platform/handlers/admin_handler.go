package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/austiecodes/dws/internal/platform/services"
)

func ListUsersAdmin(c *gin.Context) {
	adminUser, ok := requireAdminUser(c)
	if !ok {
		return
	}

	users, err := services.UsersService.List(c.Request.Context(), adminUser.ID)
	if err != nil {
		switch err {
		case services.ErrForbiddenUserManagement:
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin permission required"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	responses := make([]userJSON, 0, len(users))
	for i := range users {
		user := users[i]
		responses = append(responses, newUserJSON(&user))
	}

	c.JSON(http.StatusOK, gin.H{"users": responses})
}

type updateUserRoleRequest struct {
	IsAdmin *bool `json:"is_admin" binding:"required"`
}

func UpdateUserAdminFlag(c *gin.Context) {
	adminUser, ok := requireAdminUser(c)
	if !ok {
		return
	}

	idParam := c.Param("id")
	if idParam == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
		return
	}

	targetID64, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req updateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.IsAdmin == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "is_admin is required"})
		return
	}

	err = services.UsersService.SetAdmin(c.Request.Context(), adminUser.ID, uint(targetID64), *req.IsAdmin)
	if err != nil {
		switch err {
		case services.ErrForbiddenUserManagement:
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin permission required"})
			return
		case services.ErrUserNotFound:
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		case services.ErrLastAdmin:
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "cannot remove the last admin"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "user role updated"})
}
