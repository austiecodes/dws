package image

import "github.com/gin-gonic/gin"

type ImageRouter struct{}

func NewRouter() *ImageRouter {
	return &ImageRouter{}
}

func (r *ImageRouter) RegisterRoutes(group *gin.RouterGroup) {
	group.Group("/images")
}
