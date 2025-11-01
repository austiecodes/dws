package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	libconfig "github.com/austiecodes/dws/internal/lib/config"
	libdb "github.com/austiecodes/dws/internal/lib/db"
	"github.com/austiecodes/dws/internal/scheduler/grpcserver"
)

func main() {
	cfg, err := libconfig.LoadScheduler("")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if _, err := libdb.Init(cfg.Database); err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer libdb.Close()

	server := grpcserver.NewServer(*cfg)
	if err := server.Start(); err != nil {
		log.Fatalf("start scheduler rpc: %v", err)
	}
	defer server.Stop(context.Background())

	log.Println("[scheduler] service started")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("[scheduler] shutting down")
}
