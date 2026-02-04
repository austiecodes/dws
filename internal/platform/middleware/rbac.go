package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	libdb "github.com/austiecodes/dws/internal/lib/db"
	"github.com/austiecodes/dws/internal/lib/rbac"
	"github.com/austiecodes/dws/internal/platform/services"
)

const (
	sessionUserKey  = "user_id"
	contextUserKey  = "current_user"
	contextUserRole = "current_user_role"
)

// sessionUserID extracts the user ID from the session cookie.
func sessionUserID(c *gin.Context) uint {
	session := sessions.Default(c)
	id := session.Get(sessionUserKey)
	if id == nil {
		return 0
	}

	switch v := id.(type) {
	case uint:
		return v
	case int:
		if v < 0 {
			return 0
		}
		return uint(v)
	case int64:
		if v < 0 {
			return 0
		}
		return uint(v)
	case float64:
		if v < 0 {
			return 0
		}
		return uint(v)
	default:
		return 0
	}
}

// getCurrentUser retrieves the current user from context or database.
func getCurrentUser(c *gin.Context) (*libdb.User, error) {
	if val, exists := c.Get(contextUserKey); exists {
		if user, ok := val.(*libdb.User); ok {
			return user, nil
		}
	}

	userID := sessionUserID(c)
	if userID == 0 {
		return nil, nil
	}

	user, err := services.Auth.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		return nil, err
	}
	if user != nil {
		c.Set(contextUserKey, user)
	}
	return user, nil
}

// RequireAuth ensures the requester is authenticated.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := sessionUserID(c)
		if userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}
		c.Next()
	}
}

// RequireRole ensures the requester has at least the specified role.
// Role hierarchy: super_admin > admin > user
func RequireRole(minRole string) gin.HandlerFunc {
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

		if !rbac.HasRole(user.Role, minRole) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "permission denied"})
			return
		}

		c.Set(contextUserRole, user.Role)
		c.Next()
	}
}

// RequirePermission checks if the user has a specific permission via Casbin.
func RequirePermission(resource, action string) gin.HandlerFunc {
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

		if !rbac.HasPermissionByUserRole(user.Role, resource, action) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "permission denied"})
			return
		}

		c.Set(contextUserKey, user)
		c.Set(contextUserRole, user.Role)
		c.Next()
	}
}

// GetCurrentUser returns the current user from context (set by auth middleware).
func GetCurrentUser(c *gin.Context) *libdb.User {
	if val, exists := c.Get(contextUserKey); exists {
		if user, ok := val.(*libdb.User); ok {
			return user
		}
	}
	return nil
}

// GetCurrentUserID returns the current user's ID from session.
func GetCurrentUserID(c *gin.Context) uint {
	return sessionUserID(c)
}

// GetCurrentUserRole returns the current user's role from context.
func GetCurrentUserRole(c *gin.Context) string {
	if val, exists := c.Get(contextUserRole); exists {
		if role, ok := val.(string); ok {
			return role
		}
	}
	if user := GetCurrentUser(c); user != nil {
		return user.Role
	}
	return ""
}
