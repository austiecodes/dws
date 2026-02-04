package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	libconfig "github.com/austiecodes/dws/internal/lib/config"
	libdb "github.com/austiecodes/dws/internal/lib/db"
	libdocker "github.com/austiecodes/dws/internal/lib/docker"
	"github.com/austiecodes/dws/internal/worker"
	"github.com/austiecodes/dws/internal/worker/grpcserver"
	"github.com/austiecodes/dws/internal/worker/schedulerclient"
)

func main() {
	cfg, err := libconfig.LoadWorker("")
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

	executor := worker.NewExecutor(cfg.Worker.ID)

	controlServer := grpcserver.NewServer(cfg.Worker)
	if err := controlServer.Start(); err != nil {
		log.Fatalf("start worker control rpc: %v", err)
	}
	defer controlServer.Stop(context.Background())

	client, err := schedulerclient.New(cfg.Worker, executor)
	if err != nil {
		log.Fatalf("connect to scheduler: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client.Start(ctx)

	timeoutWatcher := worker.NewTimeoutWatcher()
	timeoutWatcher.Start()
	defer timeoutWatcher.Stop()

	log.Println("[worker] service started")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	// Try to notify scheduler we're going offline before stopping background loops
	offCtx, offCancel := context.WithTimeout(context.Background(), 3*time.Second)
	_ = client.SendOfflineOnce(offCtx)
	offCancel()
	cancel()

	log.Println("[worker] shutting down")
}
