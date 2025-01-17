package types

// GPUMetrics 表示GPU的监控指标
type GPUMetrics struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	GPUUtilization    uint32 `json:"gpuUtilization"`
	MemoryUtilization uint32 `json:"memoryUtilization"`
	Temperature       uint32 `json:"temperature"`
	MemoryTotal       int    `json:"memoryTotal"`
	MemmoryFree       int    `json:"memoryFree"`
	MemoryUsed        int    `json:"memoryUsed"`
	PowerUsage        uint32 `json:"powerUsage"`
}
