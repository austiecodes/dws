package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	libdb "github.com/austiecodes/dws/internal/lib/db"
)

func TaskCreate(ctx context.Context, db *gorm.DB, task *libdb.Task) error {
	return dbWithContext(ctx, db).Create(task).Error
}

func TaskGetByID(ctx context.Context, db *gorm.DB, id uint, preloadContainer bool) (*libdb.Task, error) {
	var task libdb.Task
	query := dbWithContext(ctx, db).Where("id = ?", id)
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

func TaskListByUser(ctx context.Context, db *gorm.DB, userID uint, preloadContainer bool) ([]libdb.Task, error) {
	var tasks []libdb.Task
	query := dbWithContext(ctx, db).
		Where("user_id = ?", userID).
		Order("created_at DESC")
	if preloadContainer {
		query = query.Preload("Container")
	}
	if err := query.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func TaskListPending(ctx context.Context, db *gorm.DB) ([]libdb.Task, error) {
	var tasks []libdb.Task
	err := dbWithContext(ctx, db).
		Where("status = ?", libdb.TaskStatusPending).
		Order("priority DESC, created_at ASC").
		Preload("Container").
		Find(&tasks).Error
	return tasks, err
}

func TaskListPendingByType(ctx context.Context, db *gorm.DB, taskType libdb.TaskType, limit int) ([]libdb.Task, error) {
	var tasks []libdb.Task
	err := dbWithContext(ctx, db).
		Where("status = ? AND task_type = ?", libdb.TaskStatusPending, taskType).
		Order("priority DESC, created_at ASC").
		Limit(limit).
		Preload("Container").
		Find(&tasks).Error
	return tasks, err
}

func TaskListPendingForWorker(ctx context.Context, db *gorm.DB, workerID string, limit int) ([]libdb.Task, error) {
	var tasks []libdb.Task
	err := dbWithContext(ctx, db).
		Joins("JOIN containers ON containers.id = tasks.container_id").
		Where("tasks.status = ? AND containers.worker_id = ?", libdb.TaskStatusPending, workerID).
		Order("tasks.priority DESC, tasks.created_at ASC").
		Limit(limit).
		Preload("Container").
		Find(&tasks).Error
	return tasks, err
}

func TaskCountRunningByType(ctx context.Context, db *gorm.DB, taskType libdb.TaskType) (int64, error) {
	var count int64
	err := dbWithContext(ctx, db).
		Model(&libdb.Task{}).
		Where("status = ? AND task_type = ?", libdb.TaskStatusRunning, taskType).
		Count(&count).Error
	return count, err
}

func TaskGetRunningCountsByType(ctx context.Context, db *gorm.DB) (map[libdb.TaskType]int, error) {
	type CountResult struct {
		TaskType string
		Count    int
	}

	var results []CountResult
	err := dbWithContext(ctx, db).
		Model(&libdb.Task{}).
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

func TaskGetPendingCountsByType(ctx context.Context, db *gorm.DB) (map[libdb.TaskType]int, error) {
	type CountResult struct {
		TaskType string
		Count    int
	}

	var results []CountResult
	err := dbWithContext(ctx, db).
		Model(&libdb.Task{}).
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

func TaskListRunning(ctx context.Context, db *gorm.DB) ([]libdb.Task, error) {
	var tasks []libdb.Task
	err := dbWithContext(ctx, db).
		Where("status = ?", libdb.TaskStatusRunning).
		Preload("Container").
		Find(&tasks).Error
	return tasks, err
}

func TaskUpdateStatus(ctx context.Context, db *gorm.DB, id uint, status libdb.TaskStatus) error {
	dbctx := dbWithContext(ctx, db)
	now := time.Now()
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": now,
	}

	if status == libdb.TaskStatusRunning {
		updates["started_at"] = now
	}

	if status == libdb.TaskStatusCompleted || status == libdb.TaskStatusFailed || status == libdb.TaskStatusKilled {
		updates["completed_at"] = now
	}

	return dbctx.Model(&libdb.Task{}).Where("id = ?", id).Updates(updates).Error
}

func TaskTransitionStatus(ctx context.Context, db *gorm.DB, id uint, from, to libdb.TaskStatus) (bool, error) {
	dbctx := dbWithContext(ctx, db)
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

	result := dbctx.Model(&libdb.Task{}).
		Where("id = ? AND status = ?", id, from).
		Updates(updates)
	return result.RowsAffected > 0, result.Error
}

func TaskUpdateResult(ctx context.Context, db *gorm.DB, id uint, output string, exitCode int) error {
	return dbWithContext(ctx, db).Model(&libdb.Task{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"output":    output,
			"exit_code": exitCode,
		}).Error
}

func TaskFindTimedOut(ctx context.Context, db *gorm.DB, timeout time.Duration) ([]libdb.Task, error) {
	var tasks []libdb.Task
	cutoff := time.Now().Add(-timeout)
	err := dbWithContext(ctx, db).
		Where("status = ? AND started_at < ?", libdb.TaskStatusRunning, cutoff).
		Preload("Container").
		Find(&tasks).Error
	return tasks, err
}

func TaskCancelByUser(ctx context.Context, db *gorm.DB, taskID, userID uint) error {
	result := dbWithContext(ctx, db).Model(&libdb.Task{}).
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

func TaskCancelByID(ctx context.Context, db *gorm.DB, taskID uint) error {
	result := dbWithContext(ctx, db).Model(&libdb.Task{}).
		Where("id = ? AND status IN ?", taskID, []libdb.TaskStatus{
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
