package scheduler

import (
	"context"
	"log"
	"time"

	libdb "github.com/austiecodes/dws/internal/lib/db"
	"github.com/austiecodes/dws/internal/platform/repository"
)

const (
	// MaxConcurrentCPU defines the maximum number of concurrent CPU tasks.
	MaxConcurrentCPU = 3
	// MaxConcurrentGPU defines the maximum number of concurrent GPU tasks (exclusive).
	MaxConcurrentGPU = 1
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

	// Get current running task counts by type
	runningCounts, err := repository.TaskGetRunningCountsByType(ctx, nil)
	if err != nil {
		log.Printf("[scheduler] failed to get running counts: %v", err)
		return
	}

	runningCPU := runningCounts[libdb.TaskTypeCPU]
	runningGPU := runningCounts[libdb.TaskTypeGPU]

	// Calculate available slots
	availableCPU := MaxConcurrentCPU - runningCPU
	availableGPU := MaxConcurrentGPU - runningGPU

	log.Printf("[scheduler] running: CPU=%d/%d, GPU=%d/%d", runningCPU, MaxConcurrentCPU, runningGPU, MaxConcurrentGPU)

	if availableCPU <= 0 && availableGPU <= 0 {
		return // No available slots
	}

	// Dispatch CPU tasks if slots available
	if availableCPU > 0 {
		d.dispatchByType(ctx, libdb.TaskTypeCPU, availableCPU)
	}

	// Dispatch GPU tasks if slots available
	if availableGPU > 0 {
		d.dispatchByType(ctx, libdb.TaskTypeGPU, availableGPU)
	}
}

func (d *Dispatcher) dispatchByType(ctx context.Context, taskType libdb.TaskType, limit int) {
	pending, err := repository.TaskListPendingByType(ctx, nil, taskType, limit)
	if err != nil {
		log.Printf("[scheduler] failed to list pending %s tasks: %v", taskType, err)
		return
	}

	if len(pending) == 0 {
		return
	}

	log.Printf("[scheduler] dispatching %d %s task(s)", len(pending), taskType)

	for _, task := range pending {
		if err := d.markAsRunning(ctx, task.ID); err != nil {
			log.Printf("[scheduler] failed to dispatch %s task %d: %v", taskType, task.ID, err)
		} else {
			log.Printf("[scheduler] dispatched %s task %d (container=%d, priority=%d)", taskType, task.ID, task.ContainerID, task.Priority)
		}
	}
}

func (d *Dispatcher) markAsRunning(ctx context.Context, taskID uint) error {
	return repository.TaskUpdateStatus(ctx, nil, taskID, libdb.TaskStatusRunning)
}
