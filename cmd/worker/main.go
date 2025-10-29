package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	libconfig "github.com/austiecodes/dws/internal/lib/config"
	libdb "github.com/austiecodes/dws/internal/lib/db"
	libdocker "github.com/austiecodes/dws/internal/lib/docker"
	"github.com/austiecodes/dws/internal/worker"
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

	// Process running tasks every 3 seconds
	executor := worker.NewExecutor(3 * time.Second)
	executor.Start()
	defer executor.Stop()

	// Watch for task timeouts
	timeoutWatcher := worker.NewTimeoutWatcher()
	timeoutWatcher.Start()
	defer timeoutWatcher.Stop()

	log.Println("[worker] service started")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("[worker] shutting down")
}
