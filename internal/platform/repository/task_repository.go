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
