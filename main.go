package main

import (
	"github.com/austiecodes/dws/routes"
	"github.com/austiecodes/dws/start"
	"github.com/gin-gonic/gin"
)

func main() {
	// TODO: add config parser here
	// the app init accroding to app.toml
	// bootstrap
	start.MustInit()

	// init http server
	r := gin.Default()
	routes.SetupRoutes(r)
	r.Run(":8989")
}
