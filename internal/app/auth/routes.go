package auth

import "github.com/gin-gonic/gin"

type AuthRouter struct{}

func NewRouter() *AuthRouter {
	return &AuthRouter{}
}

func (r *AuthRouter) RegisterRoutes(group *gin.RouterGroup) {
	ag := group.Group("/auth")
	{
		ag.POST("/login", Login)
	}
}
