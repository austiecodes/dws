package routes

import (
	"net/http"

	"github.com/austiecodes/dws/internal/app/auth"
	"github.com/austiecodes/dws/internal/app/container"
	"github.com/austiecodes/dws/lib/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		setToolsRouters(v1)
		setupAuthRouters(v1)
		// setupGPURouters(v1)
		setupContainerRouters(v1)
		// setupImageRouters(v1)
	}
}

func setToolsRouters(r *gin.RouterGroup) {
	r.GET("/alive", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func setupAuthRouters(r *gin.RouterGroup) {
	authRouters := r.Group("/auth")
	{
		authRouters.POST("/login", auth.Login)
	}
}

// func setupGPURouters(r *gin.RouterGroup) {
// 	gpuRouters := r.Group("/gpu")
// 	gpuRouters.Use(middleware.AuthMiddleware())
// 	{
// 		gpuRouters.GET("/status", gpu.GetGPUStatus)
// 	}
// }

func setupContainerRouters(r *gin.RouterGroup) {
	containerRouters := r.Group("/containers")
	containerRouters.Use(middleware.AuthMiddleware())
	{
		containerRouters.GET("/list", container.ListContainerController)
		containerRouters.GET("/running", container.ListRunningContainerController)
		containerRouters.GET("/start", container.StartContainerController)
		containerRouters.POST("/stop", container.StopContainerController)
		containerRouters.POST("/create", container.CreateContaineController)
		containerRouters.DELETE("/remove", container.RemoveContainerController)
	}
}

// func setupImageRouters(r *gin.RouterGroup) {
// 	imageRouters := r.Group("/images")
// 	imageRouters.Use(middleware.AuthMiddleware())
// 	{
// 		imageRouters.GET("/list", images.ListImages)
// 	}
// }
