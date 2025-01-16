package routes

import (
	"github.com/austiecodes/dws/controllers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// GPU routes

	v1 := r.Group("/api/v1")
	{
		gpus := v1.Group("/gpu")
		{
			gpus.GET("/status", func(c *gin.Context) {
				data, err := controllers.GetGPUMetricsService()
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, data)
			})
		}
	}
}
