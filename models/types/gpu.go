package types

import (
	"time"
)

// GPU 表示一个GPU设备
type GPU struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Model       string    `json:"model"`
	MemoryUsed  int       `json:"memory_used"`  // in MB
	MemoryTotal int       `json:"memory_total"` // in MB
	Status      string    `json:"status"`       // available, in-use, maintenance
	Temperature float64   `json:"temperature"`
	Usage       float64   `json:"usage"`      // in percentage
	PowerDraw   float64   `json:"power_draw"` // in watts
	UserID      *uint     `json:"user_id"`    // who is using the GPU
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GPUMetrics 表示GPU的监控指标
type GPUMetrics struct {
	GPUID       uint    `json:"gpu_id"`
	Name        string  `json:"name"` // 新增：GPU名称
	Temperature float64 `json:"temperature"`
	MemoryUsed  int     `json:"memory_used"`
	MemoryTotal int     `json:"memory_total"`
	Utilization float64 `json:"utilization"`
	PowerDraw   float64 `json:"power_draw"`
}
