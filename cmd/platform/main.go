package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	libconfig "github.com/austiecodes/dws/internal/lib/config"
	libdb "github.com/austiecodes/dws/internal/lib/db"
	libdocker "github.com/austiecodes/dws/internal/lib/docker"
	"github.com/austiecodes/dws/internal/platform/router"
	"github.com/austiecodes/dws/internal/platform/services"
)

func main() {
	cfg, err := libconfig.Load("")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if _, err := libdb.Init(cfg.Database); err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer libdb.Close()

	if _, err := libdocker.Init(cfg.Docker); err != nil {
		log.Fatalf("init docker: %v", err)
	}

	engine := gin.New()
	if err := router.Setup(engine, cfg); err != nil {
		log.Fatalf("init engine: %v", err)
	}

	// 启动容器状态同步服务（每 5 分钟同步一次）
	syncService := services.NewSyncService(5 * time.Minute)
	syncService.Start()
	defer syncService.Stop()

	addr := fmt.Sprintf(":%d", cfg.App.Port)
	log.Printf("Starting server on %s", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
