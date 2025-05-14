package router

import (
	"github.com/gin-gonic/gin"
)

// Router interface defines the contract for all routers in the application
type Router interface {
	RegisterRoutes(group *gin.RouterGroup)
}

// InitRouter initializes all application routers
func InitRouter(routers []Router) *gin.Engine {
	r := gin.Default()
	api := r.Group("/api")

	// Register each router
	for _, router := range routers {
		router.RegisterRoutes(api)
	}

	return r
}
