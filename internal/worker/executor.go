package worker

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/austiecodes/dws/internal/lib/apis/schedulerpb"
	libdb "github.com/austiecodes/dws/internal/lib/db"
	libdocker "github.com/austiecodes/dws/internal/lib/docker"
)

// Executor runs task assignments delivered by the scheduler.
type Executor struct {
	workerID string
	docker   *libdocker.Manager

	executing sync.Map // taskID -> time started
}

// NewExecutor constructs an executor bound to a worker identity.
func NewExecutor(workerID string) *Executor {
	return &Executor{
		workerID: workerID,
		docker:   libdocker.MustInstance(),
	}
}

// ActiveTaskCount returns the number of tasks currently executing.
func (e *Executor) ActiveTaskCount() int32 {
	count := int32(0)
	e.executing.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// Execute runs the provided assignment and returns the result payload.
func (e *Executor) Execute(ctx context.Context, assignment *schedulerpb.TaskAssignment) *schedulerpb.TaskResult {
	result := &schedulerpb.TaskResult{
		TaskID:   assignment.TaskID,
		WorkerID: e.workerID,
		Status:   string(libdb.TaskStatusFailed),
	}

	if assignment.ContainerID == "" {
		result.Error = "missing container id"
		return result
	}

	if assignment.Command == "" {
		result.Error = "empty command"
		return result
	}

	if _, loaded := e.executing.LoadOrStore(assignment.TaskID, time.Now()); loaded {
		result.Error = "task already executing"
		return result
	}
	defer e.executing.Delete(assignment.TaskID)

	log.Printf("[worker] executing task %d command=%s", assignment.TaskID, assignment.Command)

	cmd := strings.Fields(assignment.Command)
	if len(cmd) == 0 {
		result.Error = "invalid command"
		return result
	}

	execResult := e.docker.ExecCommand(ctx, assignment.ContainerID, cmd)
	if execResult.Error != nil {
		log.Printf("[worker] task %d failed: %v", assignment.TaskID, execResult.Error)
		result.Error = execResult.Error.Error()
		result.Output = execResult.Output
		result.ExitCode = int32(-1)
		return result
	}

	result.Output = execResult.Output
	result.ExitCode = int32(execResult.ExitCode)
	if execResult.ExitCode == 0 {
		result.Status = string(libdb.TaskStatusCompleted)
	} else {
		result.Status = string(libdb.TaskStatusFailed)
	}

	log.Printf("[worker] task %d completed exit=%d", assignment.TaskID, execResult.ExitCode)
	return result
}
