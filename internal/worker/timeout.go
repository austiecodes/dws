package worker

import (
	"context"
	"log"
	"time"

	libdb "github.com/austiecodes/dws/internal/lib/db"
	"github.com/austiecodes/dws/internal/platform/repository"
)

const (
	MaxTaskDuration      = 30 * time.Minute
	NotificationDuration = 5 * time.Minute // Mock notification 5 minutes before kill
	TimeoutCheckInterval = 1 * time.Minute
)

type TimeoutWatcher struct {
	stopCh chan struct{}
}

func NewTimeoutWatcher() *TimeoutWatcher {
	return &TimeoutWatcher{
		stopCh: make(chan struct{}),
	}
}

// Start begins the timeout watcher loop.
func (w *TimeoutWatcher) Start() {
	go w.run()
}

// Stop gracefully stops the timeout watcher.
func (w *TimeoutWatcher) Stop() {
	close(w.stopCh)
}

func (w *TimeoutWatcher) run() {
	ticker := time.NewTicker(TimeoutCheckInterval)
	defer ticker.Stop()

	log.Println("[worker] timeout watcher started")

	for {
		select {
		case <-ticker.C:
			w.checkTimeouts()
		case <-w.stopCh:
			log.Println("[worker] timeout watcher stopped")
			return
		}
	}
}

func (w *TimeoutWatcher) checkTimeouts() {
	ctx := context.Background()

	// Find tasks that have exceeded the 30-minute hard limit
	timedOut, err := repository.Tasks.FindTimedOutTasks(MaxTaskDuration)
	if err != nil {
		log.Printf("[worker] failed to find timed-out tasks: %v", err)
		return
	}

	for _, task := range timedOut {
		log.Printf("[worker] killing task %d (exceeded 30 minutes)", task.ID)
		w.killTask(ctx, task)
	}

	// Mock notification: check for tasks approaching timeout
	approaching, err := repository.Tasks.FindTimedOutTasks(MaxTaskDuration - NotificationDuration)
	if err != nil {
		log.Printf("[worker] failed to find approaching-timeout tasks: %v", err)
		return
	}

	for _, task := range approaching {
		if task.StartedAt == nil {
			continue
		}
		elapsed := time.Since(*task.StartedAt)
		remaining := MaxTaskDuration - elapsed

		// Only notify if within notification window and still running
		if remaining > 0 && remaining <= NotificationDuration {
			w.notifyUser(task, remaining)
		}
	}
}

func (w *TimeoutWatcher) killTask(ctx context.Context, task libdb.Task) {
	output := "Task killed by system: exceeded maximum runtime of 30 minutes"
	if err := repository.Tasks.UpdateResult(task.ID, output, -1); err != nil {
		log.Printf("[worker] failed to update task %d result: %v", task.ID, err)
	}
	if err := repository.Tasks.UpdateStatus(task.ID, libdb.TaskStatusKilled); err != nil {
		log.Printf("[worker] failed to kill task %d: %v", task.ID, err)
	}
}

func (w *TimeoutWatcher) notifyUser(task libdb.Task, remaining time.Duration) {
	// Mock notification: just log it
	// In production, this would send email/SMS/browser notification
	log.Printf("[worker] MOCK NOTIFICATION: Task %d (user %d) has %.0f minutes remaining before forced termination",
		task.ID, task.UserID, remaining.Minutes())
}
