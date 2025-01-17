package routes

import (
	"net/http"

	"github.com/austiecodes/dws/controllers/containers"
	"github.com/austiecodes/dws/controllers/gpu"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		setToolsRouters(v1)
		setupGPURouters(v1)
		setupContainerRouters(v1)
	}
}

func setToolsRouters(r *gin.RouterGroup) {
	r.GET("/alive", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func setupGPURouters(r *gin.RouterGroup) {
	gpuRouters := r.Group("/gpu")
	{
		gpuRouters.GET("/status", gpu.GetStatusController)
	}
}

func setupContainerRouters(r *gin.RouterGroup) {
	containerRouters := r.Group("/containers")
	{
		containerRouters.GET("/running", containers.GetRuningContainers)
	}
}
