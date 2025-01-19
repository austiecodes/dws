package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get session
		session := sessions.Default(c)
		// check user field
		user := session.Get("uuid")
		if user == nil {
			c.JSON(401, gin.H{
				"message": "Not logged in",
			})
			c.Abort()
			return
		}

		c.Set("uuid", user)
		c.Next()
	}
}
