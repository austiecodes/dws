package handlers

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

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
