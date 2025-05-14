package server

import "github.com/gin-gonic/gin"

// Router interface defines the contract for all routers in the application
type Router interface {
	RegisterRoutes(group *gin.RouterGroup)
}

// RegisterRouters registers all routers with the given gin engine
func RegisterRouters(e *gin.Engine, routers ...Router) {
	// Create API version group
	v1 := e.Group("/api/v1")

	// Register each router
	for _, router := range routers {
		router.RegisterRoutes(v1)
	}
}
