package main

import (
	"github.com/austiecodes/dws/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// 设置路由
	routes.SetupRoutes(r)

	// 启动服务器
	r.Run(":8080")
}
