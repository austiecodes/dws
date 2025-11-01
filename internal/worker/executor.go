package worker

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	libdb "github.com/austiecodes/dws/internal/lib/db"
	libdocker "github.com/austiecodes/dws/internal/lib/docker"
	"github.com/austiecodes/dws/internal/platform/repository"
)

type Executor struct {
	interval       time.Duration
	stopCh         chan struct{}
	executingTasks sync.Map // taskID -> true, prevents duplicate execution
}

func NewExecutor(interval time.Duration) *Executor {
	return &Executor{
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the executor loop that picks up running tasks and executes them.
func (e *Executor) Start() {
	go e.run()
}

// Stop gracefully stops the executor.
func (e *Executor) Stop() {
	close(e.stopCh)
}

func (e *Executor) run() {
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	log.Println("[worker] executor started")

	for {
		select {
		case <-ticker.C:
			e.processRunningTasks()
		case <-e.stopCh:
			log.Println("[worker] executor stopped")
			return
		}
	}
}

func (e *Executor) processRunningTasks() {
	ctx := context.Background()

	running, err := repository.Tasks.ListRunning()
	if err != nil {
		log.Printf("[worker] failed to list running tasks: %v", err)
		return
	}

	if len(running) == 0 {
		return
	}

	dispatched := 0
	for _, task := range running {
		// Skip if already executing
		if _, exists := e.executingTasks.Load(task.ID); exists {
			continue
		}

		// Mark as executing and launch goroutine
		e.executingTasks.Store(task.ID, true)
		dispatched++
		go e.executeTask(ctx, task)
	}

	if dispatched > 0 {
		log.Printf("[worker] dispatched %d task(s) for execution", dispatched)
	}
}

func (e *Executor) executeTask(ctx context.Context, task libdb.Task) {
	// Ensure we always remove the executing flag when done
	defer e.executingTasks.Delete(task.ID)

	log.Printf("[worker] executing task %d: container=%s command=%s", task.ID, task.Container.ContainerID, task.Command)

	mgr := libdocker.MustInstance()

	// Parse command (simple space-split, could be improved)
	cmd := strings.Fields(task.Command)
	if len(cmd) == 0 {
		e.failTask(task.ID, "empty command", -1)
		return
	}

	// Execute command in container
	result := mgr.ExecCommand(ctx, task.Container.ContainerID, cmd)

	if result.Error != nil {
		log.Printf("[worker] task %d failed: %v", task.ID, result.Error)
		e.failTask(task.ID, result.Error.Error(), -1)
		return
	}

	log.Printf("[worker] task %d completed with exit code %d", task.ID, result.ExitCode)

	// Update task result
	if err := repository.Tasks.UpdateResult(task.ID, result.Output, result.ExitCode); err != nil {
		log.Printf("[worker] failed to update task %d result: %v", task.ID, err)
	}

	// Mark as completed or failed based on exit code
	status := libdb.TaskStatusCompleted
	if result.ExitCode != 0 {
		status = libdb.TaskStatusFailed
	}
	if err := repository.Tasks.UpdateStatus(task.ID, status); err != nil {
		log.Printf("[worker] failed to update task %d status: %v", task.ID, err)
	}
}

func (e *Executor) failTask(taskID uint, output string, exitCode int) {
	if err := repository.Tasks.UpdateResult(taskID, output, exitCode); err != nil {
		log.Printf("[worker] failed to update task %d result: %v", taskID, err)
	}
	if err := repository.Tasks.UpdateStatus(taskID, libdb.TaskStatusFailed); err != nil {
		log.Printf("[worker] failed to update task %d status: %v", taskID, err)
	}
}
