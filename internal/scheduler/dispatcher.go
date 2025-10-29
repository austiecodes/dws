package scheduler

import (
	"context"
	"log"
	"time"

	libdb "github.com/austiecodes/dws/internal/lib/db"
	"github.com/austiecodes/dws/internal/platform/repository"
)

type Dispatcher struct {
	interval time.Duration
	stopCh   chan struct{}
}

func NewDispatcher(interval time.Duration) *Dispatcher {
	return &Dispatcher{
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the dispatch loop that promotes pending tasks to running.
func (d *Dispatcher) Start() {
	go d.run()
}

// Stop gracefully stops the dispatcher.
func (d *Dispatcher) Stop() {
	close(d.stopCh)
}

func (d *Dispatcher) run() {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	log.Println("[scheduler] dispatcher started")

	for {
		select {
		case <-ticker.C:
			d.dispatchPendingTasks()
		case <-d.stopCh:
			log.Println("[scheduler] dispatcher stopped")
			return
		}
	}
}

func (d *Dispatcher) dispatchPendingTasks() {
	ctx := context.Background()

	// Fetch all pending tasks (already ordered by priority DESC, created_at ASC)
	pending, err := repository.Tasks.ListPending()
	if err != nil {
		log.Printf("[scheduler] failed to list pending tasks: %v", err)
		return
	}

	if len(pending) == 0 {
		return
	}

	log.Printf("[scheduler] found %d pending task(s)", len(pending))

	// Simple strategy: mark all pending as running immediately
	// In a real system, you might check worker capacity here
	for _, task := range pending {
		if err := d.markAsRunning(ctx, task.ID); err != nil {
			log.Printf("[scheduler] failed to mark task %d as running: %v", task.ID, err)
		} else {
			log.Printf("[scheduler] dispatched task %d (container=%d, priority=%d)", task.ID, task.ContainerID, task.Priority)
		}
	}
}

func (d *Dispatcher) markAsRunning(ctx context.Context, taskID uint) error {
	return repository.Tasks.UpdateStatus(taskID, libdb.TaskStatusRunning)
}
