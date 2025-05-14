package task

import "github.com/gin-gonic/gin"

type TaskRouter struct{}

func NewRouter() *TaskRouter {
	return &TaskRouter{}
}

func (r *TaskRouter) RegisterRoutes(group *gin.RouterGroup) {
	group.Group("/tasks")
}
