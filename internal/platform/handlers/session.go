package handlers

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	libdb "github.com/austiecodes/dws/internal/lib/db"
	"github.com/austiecodes/dws/internal/platform/services"
)

const contextUserKey = "current_user"

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

func requireUser(c *gin.Context) (uint, bool) {
	userID := sessionUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return 0, false
	}
	return userID, true
}

// RequireAuthMiddleware enforces authentication at the route group level
func RequireAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := sessionUserID(c)
		if userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}
		c.Next()
	}
}

// RequireAdminMiddleware ensures the requester is authenticated and has admin privileges.
func RequireAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := sessionUserID(c)
		if userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		user, err := services.Auth.GetUserByID(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
			return
		}
		if user == nil || !user.IsAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin permission required"})
			return
		}
		c.Set(contextUserKey, user)
		c.Next()
	}
}

func requireAdminUser(c *gin.Context) (*libdb.User, bool) {
	value, exists := c.Get(contextUserKey)
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin permission required"})
		return nil, false
	}
	user, ok := value.(*libdb.User)
	if !ok || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid admin context"})
		return nil, false
	}
	return user, true
}
