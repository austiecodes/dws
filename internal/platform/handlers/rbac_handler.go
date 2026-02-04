package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/austiecodes/dws/internal/lib/rbac"
	"github.com/austiecodes/dws/internal/platform/middleware"
	"github.com/austiecodes/dws/internal/platform/services"
)

// ListRoles returns all available roles.
// GET /api/v1/rbac/roles
func ListRoles(c *gin.Context) {
	roles := services.RBAC.ListRoles()
	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// GetMyPermissions returns the current user's effective permissions.
// GET /api/v1/rbac/permissions/me
func GetMyPermissions(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	permissions, err := services.RBAC.GetUserPermissions(c.Request.Context(), userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get user role
	role, err := services.RBAC.GetUserRole(c.Request.Context(), userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Format permissions as resource:action pairs
	formattedPerms := make([]string, 0, len(permissions))
	for _, p := range permissions {
		if len(p) >= 3 {
			formattedPerms = append(formattedPerms, p[1]+":"+p[2])
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"role":        role,
		"permissions": formattedPerms,
	})
}

type assignRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

// AssignUserRole assigns a role to a user.
// POST /api/v1/admin/users/:id/role
func AssignUserRole(c *gin.Context) {
	requesterID := middleware.GetCurrentUserID(c)
	if requesterID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
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

	var req assignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "role is required"})
		return
	}

	if !rbac.IsValidRole(req.Role) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":        "invalid role",
			"valid_roles":  rbac.ValidRoles,
		})
		return
	}

	err = services.RBAC.AssignRole(c.Request.Context(), requesterID, uint(targetID64), req.Role)
	if err != nil {
		switch err {
		case services.ErrCannotAssignRoles:
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "only super_admin can assign roles"})
			return
		case services.ErrUserNotFound:
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		case services.ErrLastSuperAdmin:
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "cannot remove the last super_admin"})
			return
		case services.ErrCannotDowngrade:
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "cannot downgrade your own role"})
			return
		case services.ErrInvalidRole:
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "role assigned successfully"})
}

// GetUserPermissions returns a specific user's effective permissions.
// GET /api/v1/admin/users/:id/permissions
func GetUserPermissions(c *gin.Context) {
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

	permissions, err := services.RBAC.GetUserPermissions(c.Request.Context(), uint(targetID64))
	if err != nil {
		if err == services.ErrUserNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	role, err := services.RBAC.GetUserRole(c.Request.Context(), uint(targetID64))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Format permissions
	formattedPerms := make([]string, 0, len(permissions))
	for _, p := range permissions {
		if len(p) >= 3 {
			formattedPerms = append(formattedPerms, p[1]+":"+p[2])
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":     targetID64,
		"role":        role,
		"permissions": formattedPerms,
	})
}
