package routes

import (
	"net/http"

	"github.com/austiecodes/dws/controllers/containers"
	"github.com/austiecodes/dws/controllers/gpu"
	"github.com/austiecodes/dws/controllers/images"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		setToolsRouters(v1)
		setupGPURouters(v1)
		setupContainerRouters(v1)
		setupImageRouters(v1)
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
		gpuRouters.GET("/status", gpu.GetGPUStatus)
	}
}

func setupContainerRouters(r *gin.RouterGroup) {
	containerRouters := r.Group("/containers")
	{
		containerRouters.GET("/list", containers.ListContainers)
		containerRouters.GET("/running", containers.ListRunningContainers)
		containerRouters.GET("/start", containers.StopContainers)
		containerRouters.POST("/stop", containers.StopContainers)
		containerRouters.POST("/remove", containers.RemoveContainers)
		containerRouters.POST("/create", containers.CreateContainer)

	}
}

func setupImageRouters(r *gin.RouterGroup) {
	imageRouters := r.Group("/images")
	{
		imageRouters.GET("/list", images.ListImages)
	}
}
