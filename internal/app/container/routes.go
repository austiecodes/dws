package container

import "github.com/gin-gonic/gin"

// ContainerRouter 封装了 Controller 的注册结构体
type ContainerRouter struct{}

// NewRouter 构造函数
func NewRouter() *ContainerRouter {
	return &ContainerRouter{}
}

// RegisterRoutes 实现统一接口
func (r *ContainerRouter) RegisterRoutes(group *gin.RouterGroup) {
	cg := group.Group("/containers")
	{
		cg.POST("/list", ListContainerController)
		cg.POST("/list/running", ListRunningContainerController)
		cg.POST("/start", StartContainerController)
		cg.POST("/stop", StopContainerController)
		cg.POST("/create", CreateContaineController)
		cg.POST("/remove", RemoveContainerController)
	}
}
