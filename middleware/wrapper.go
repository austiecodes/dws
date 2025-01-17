package middleware

import (
	"github.com/gin-gonic/gin"
)

func HandlerWrapper(handler func(c *gin.Context) (any, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := handler(c) // 将 gin.Context 传递给被包裹的控制器函数
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, data)
	}
}
