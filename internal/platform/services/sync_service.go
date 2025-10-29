package services

import (
	"context"
	"log"
	"time"
)

type SyncService struct {
	interval time.Duration
	stopCh   chan struct{}
}

func NewSyncService(interval time.Duration) *SyncService {
	if interval < time.Minute {
		interval = 5 * time.Minute // 最小 5 分钟
	}
	return &SyncService{
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start 启动后台同步任务
func (s *SyncService) Start() {
	go s.run()
}

// Stop 停止后台同步任务
func (s *SyncService) Stop() {
	close(s.stopCh)
}

func (s *SyncService) run() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	log.Printf("[SyncService] Started with interval: %v", s.interval)

	// 启动时立即执行一次同步
	s.syncOnce()

	for {
		select {
		case <-ticker.C:
			s.syncOnce()
		case <-s.stopCh:
			log.Println("[SyncService] Stopped")
			return
		}
	}
}

func (s *SyncService) syncOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := Containers.SyncAllContainers(ctx); err != nil {
		log.Printf("[SyncService] Sync failed: %v", err)
	} else {
		log.Println("[SyncService] Sync completed successfully")
	}
}
