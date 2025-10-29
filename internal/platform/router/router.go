package router

import (
	"fmt"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	libconfig "github.com/austiecodes/dws/internal/lib/config"
	"github.com/austiecodes/dws/internal/platform/handlers"
	"github.com/austiecodes/dws/internal/platform/services"
)

// Setup attaches middleware and routes to the provided gin engine.
func Setup(engine *gin.Engine, cfg *libconfig.AppConfig) error {
	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("init zap logger: %w", err)
	}
	engine.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	engine.Use(ginzap.RecoveryWithZap(logger, true))

	store := buildSessionStore(cfg)
	engine.Use(sessions.Sessions(cfg.App.SessionName, store))

	services.InitContainerService(cfg.Docker)

	api := engine.Group("/api/v1")
	auth := api.Group("/auth")
	auth.POST("/register", handlers.Register)
	auth.POST("/login", handlers.Login)
	auth.POST("/logout", handlers.Logout)
	auth.GET("/me", handlers.Me)

	// All container endpoints require authentication
	containers := api.Group("/containers")
	containers.Use(handlers.RequireAuthMiddleware())
	containers.GET("", handlers.ListContainers)
	containers.POST("", handlers.CreateContainer)
	containers.POST("/:uuid/stop", handlers.StopContainer)
	containers.POST("/:uuid/start", handlers.StartContainer)
	containers.DELETE("/:uuid", handlers.DeleteContainer)
	containers.GET("/images", handlers.ListImages)

	return nil
}

func buildSessionStore(cfg *libconfig.AppConfig) sessions.Store {
	if cfg.App.AESKey != "" {
		return cookie.NewStore([]byte(cfg.App.SessionKey), []byte(cfg.App.AESKey))
	}
	return cookie.NewStore([]byte(cfg.App.SessionKey))
}
