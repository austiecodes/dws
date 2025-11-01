package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	libdb "github.com/austiecodes/dws/internal/lib/db"
)

type TaskRepository struct{}

var Tasks = &TaskRepository{}

// Create inserts a new task.
func (r *TaskRepository) Create(task *libdb.Task) error {
	conn := libdb.MustInstance()
	return conn.Create(task).Error
}

// GetByID retrieves a single task by ID, optionally preloading container info.
func (r *TaskRepository) GetByID(id uint, preloadContainer bool) (*libdb.Task, error) {
	conn := libdb.MustInstance()
	var task libdb.Task
	query := conn.Where("id = ?", id)
	if preloadContainer {
		query = query.Preload("Container")
	}
	if err := query.First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

// ListByUser returns all tasks for a given user, ordered by creation time (newest first).
func (r *TaskRepository) ListByUser(userID uint, preloadContainer bool) ([]libdb.Task, error) {
	conn := libdb.MustInstance()
	var tasks []libdb.Task
	query := conn.Where("user_id = ?", userID).Order("created_at DESC")
	if preloadContainer {
		query = query.Preload("Container")
	}
	if err := query.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// ListPending retrieves all pending tasks ordered by priority (descending) then created_at (ascending).
func (r *TaskRepository) ListPending() ([]libdb.Task, error) {
	conn := libdb.MustInstance()
	var tasks []libdb.Task
	err := conn.
		Where("status = ?", libdb.TaskStatusPending).
		Order("priority DESC, created_at ASC").
		Preload("Container").
		Find(&tasks).Error
	return tasks, err
}

// ListPendingByType retrieves pending tasks of a specific type, limited by count.
func (r *TaskRepository) ListPendingByType(taskType libdb.TaskType, limit int) ([]libdb.Task, error) {
	conn := libdb.MustInstance()
	var tasks []libdb.Task
	err := conn.
		Where("status = ? AND task_type = ?", libdb.TaskStatusPending, taskType).
		Order("priority DESC, created_at ASC").
		Limit(limit).
		Preload("Container").
		Find(&tasks).Error
	return tasks, err
}

// ListPendingForWorker retrieves pending tasks that target a specific worker.
func (r *TaskRepository) ListPendingForWorker(workerID string, limit int) ([]libdb.Task, error) {
	conn := libdb.MustInstance()
	var tasks []libdb.Task
	err := conn.
		Joins("JOIN containers ON containers.id = tasks.container_id").
		Where("tasks.status = ? AND containers.worker_id = ?", libdb.TaskStatusPending, workerID).
		Order("tasks.priority DESC, tasks.created_at ASC").
		Limit(limit).
		Preload("Container").
		Find(&tasks).Error
	return tasks, err
}

// CountRunningByType returns the count of running tasks for a specific type.
func (r *TaskRepository) CountRunningByType(taskType libdb.TaskType) (int64, error) {
	conn := libdb.MustInstance()
	var count int64
	err := conn.Model(&libdb.Task{}).
		Where("status = ? AND task_type = ?", libdb.TaskStatusRunning, taskType).
		Count(&count).Error
	return count, err
}

// GetRunningCountsByType returns a map of task counts per type for running tasks.
func (r *TaskRepository) GetRunningCountsByType() (map[libdb.TaskType]int, error) {
	conn := libdb.MustInstance()
	type CountResult struct {
		TaskType string
		Count    int
	}

	var results []CountResult
	err := conn.Model(&libdb.Task{}).
		Select("task_type, COUNT(*) as count").
		Where("status = ?", libdb.TaskStatusRunning).
		Group("task_type").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[libdb.TaskType]int)
	for _, r := range results {
		counts[libdb.TaskType(r.TaskType)] = r.Count
	}
	return counts, nil
}

// GetPendingCountsByType returns a map of task counts per type for pending tasks.
func (r *TaskRepository) GetPendingCountsByType() (map[libdb.TaskType]int, error) {
	conn := libdb.MustInstance()
	type CountResult struct {
		TaskType string
		Count    int
	}

	var results []CountResult
	err := conn.Model(&libdb.Task{}).
		Select("task_type, COUNT(*) as count").
		Where("status = ?", libdb.TaskStatusPending).
		Group("task_type").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[libdb.TaskType]int)
	for _, r := range results {
		counts[libdb.TaskType(r.TaskType)] = r.Count
	}
	return counts, nil
}

// ListRunning retrieves all currently running tasks.
func (r *TaskRepository) ListRunning() ([]libdb.Task, error) {
	conn := libdb.MustInstance()
	var tasks []libdb.Task
	err := conn.
		Where("status = ?", libdb.TaskStatusRunning).
		Preload("Container").
		Find(&tasks).Error
	return tasks, err
}

// UpdateStatus updates the status of a task.
func (r *TaskRepository) UpdateStatus(id uint, status libdb.TaskStatus) error {
	conn := libdb.MustInstance()
	now := time.Now()
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": now,
	}

	// Set started_at when transitioning to running
	if status == libdb.TaskStatusRunning {
		updates["started_at"] = now
	}

	// Set completed_at when reaching a terminal state
	if status == libdb.TaskStatusCompleted || status == libdb.TaskStatusFailed || status == libdb.TaskStatusKilled {
		updates["completed_at"] = now
	}

	return conn.Model(&libdb.Task{}).Where("id = ?", id).Updates(updates).Error
}

// TransitionStatus attempts to change the status from `from` to `to` and returns true if it succeeded.
func (r *TaskRepository) TransitionStatus(id uint, from, to libdb.TaskStatus) (bool, error) {
	conn := libdb.MustInstance()
	now := time.Now()
	updates := map[string]interface{}{
		"status":     to,
		"updated_at": now,
	}

	if to == libdb.TaskStatusRunning {
		updates["started_at"] = now
	}

	if to == libdb.TaskStatusCompleted || to == libdb.TaskStatusFailed || to == libdb.TaskStatusKilled {
		updates["completed_at"] = now
	}

	result := conn.Model(&libdb.Task{}).
		Where("id = ? AND status = ?", id, from).
		Updates(updates)
	return result.RowsAffected > 0, result.Error
}

// UpdateResult updates the output and exit code of a task.
func (r *TaskRepository) UpdateResult(id uint, output string, exitCode int) error {
	conn := libdb.MustInstance()
	return conn.Model(&libdb.Task{}).Where("id = ?", id).Updates(map[string]interface{}{
		"output":    output,
		"exit_code": exitCode,
	}).Error
}

// FindTimedOutTasks returns tasks that have been running for more than 30 minutes.
func (r *TaskRepository) FindTimedOutTasks(timeout time.Duration) ([]libdb.Task, error) {
	conn := libdb.MustInstance()
	var tasks []libdb.Task
	cutoff := time.Now().Add(-timeout)
	err := conn.
		Where("status = ? AND started_at < ?", libdb.TaskStatusRunning, cutoff).
		Preload("Container").
		Find(&tasks).Error
	return tasks, err
}

// CancelByUser allows a user to cancel their own pending or running task.
func (r *TaskRepository) CancelByUser(taskID, userID uint) error {
	conn := libdb.MustInstance()
	result := conn.Model(&libdb.Task{}).
		Where("id = ? AND user_id = ? AND status IN ?", taskID, userID, []libdb.TaskStatus{
			libdb.TaskStatusPending,
			libdb.TaskStatusRunning,
		}).
		Updates(map[string]interface{}{
			"status":       libdb.TaskStatusKilled,
			"completed_at": time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("task not found or not cancellable")
	}
	return nil
}
