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
	"github.com/austiecodes/dws/internal/lib/rbac"
	"github.com/austiecodes/dws/internal/platform/handlers"
	"github.com/austiecodes/dws/internal/platform/middleware"
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

	// Initialize RBAC
	if _, err := rbac.Init(); err != nil {
		return fmt.Errorf("init rbac: %w", err)
	}
	if cfg.RBAC.AutoSeed {
		if err := rbac.SeedDefaultPolicies(); err != nil {
			return fmt.Errorf("seed rbac policies: %w", err)
		}
	}

	services.InitContainerService(cfg.Docker)
	services.InitTaskService()
	services.InitWorkerService()

	api := engine.Group("/api/v1")
	auth := api.Group("/auth")
	auth.POST("/register", handlers.Register)
	auth.POST("/login", handlers.Login)
	auth.POST("/logout", handlers.Logout)
	auth.GET("/me", handlers.Me)

	// RBAC endpoints (authenticated users can view their permissions)
	rbacRoutes := api.Group("/rbac")
	rbacRoutes.Use(handlers.RequireAuthMiddleware())
	rbacRoutes.GET("/roles", handlers.ListRoles)
	rbacRoutes.GET("/permissions/me", handlers.GetMyPermissions)

	// All container endpoints require authentication
	containers := api.Group("/containers")
	containers.Use(handlers.RequireAuthMiddleware())
	containers.GET("", handlers.ListContainers)
	containers.POST("", handlers.CreateContainer)
	containers.GET("/images", handlers.ListImages)
	// Container management with ownership check
	containers.POST("/:uuid/stop", middleware.RequireContainerOwnershipOr("manage_any"), handlers.StopContainer)
	containers.POST("/:uuid/start", middleware.RequireContainerOwnershipOr("manage_any"), handlers.StartContainer)
	containers.DELETE("/:uuid", middleware.RequireContainerOwnershipOr("manage_any"), handlers.DeleteContainer)

	// Task endpoints
	tasks := api.Group("/tasks")
	tasks.Use(handlers.RequireAuthMiddleware())
	tasks.POST("", handlers.CreateTask)
	tasks.GET("", handlers.ListTasks)
	tasks.GET("/queue", handlers.GetQueueStatus) // Global queue status
	tasks.GET("/:id", handlers.GetTask)
	// Task management with ownership check
	tasks.POST("/:id/cancel", middleware.RequireTaskOwnershipOr("manage_any"), handlers.CancelTask)

	// Worker endpoints (list is public to authenticated users, CUD requires admin)
	workers := api.Group("/workers")
	workers.Use(handlers.RequireAuthMiddleware())
	workers.GET("", handlers.ListWorkers)
	workers.GET("/:id", handlers.GetWorker)

	// Worker admin operations - keep using RequireRole for infrastructure nodes
	workerAdmin := workers.Group("")
	workerAdmin.Use(middleware.RequireRole(rbac.RoleAdmin))
	workerAdmin.POST("", handlers.CreateWorker)
	workerAdmin.PUT("/:id", handlers.UpdateWorker)
	workerAdmin.DELETE("/:id", handlers.DeleteWorker)

	// Admin endpoints
	admin := api.Group("/admin")
	admin.Use(middleware.RequirePermission("users", "list"))
	admin.GET("/users", handlers.ListUsersAdmin)
	admin.PATCH("/users/:id", handlers.UpdateUserAdminFlag) // Deprecated: use POST /users/:id/role instead
	admin.GET("/users/:id/permissions", handlers.GetUserPermissions)
	// Role assignment requires super_admin
	admin.POST("/users/:id/role", middleware.RequirePermission("users", "assign_roles"), handlers.AssignUserRole)

	return nil
}

func buildSessionStore(cfg *libconfig.AppConfig) sessions.Store {
	if cfg.App.AESKey != "" {
		return cookie.NewStore([]byte(cfg.App.SessionKey), []byte(cfg.App.AESKey))
	}
	return cookie.NewStore([]byte(cfg.App.SessionKey))
}
