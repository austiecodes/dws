package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	libconfig "github.com/austiecodes/dws/internal/lib/config"
	libdb "github.com/austiecodes/dws/internal/lib/db"
	"github.com/austiecodes/dws/internal/scheduler"
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

	// Dispatch pending tasks every 5 seconds
	dispatcher := scheduler.NewDispatcher(5 * time.Second)
	dispatcher.Start()
	defer dispatcher.Stop()

	log.Println("[scheduler] service started")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("[scheduler] shutting down")
}
