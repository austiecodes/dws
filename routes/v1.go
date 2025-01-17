package routes

import (
	"github.com/austiecodes/dws/controllers/gpu"
	"github.com/austiecodes/dws/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// GPU routes

	v1 := r.Group("/api/v1")
	{
		gpus := v1.Group("/gpu")
		{
			gpus.GET("/status", middleware.HandlerWrapper(gpu.GetStatusController))
		}
	}
}
