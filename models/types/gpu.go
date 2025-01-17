package types

// GPUMetrics
type GPUMetrics struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	GPUUtilization    uint32 `json:"gpuUtilization"`
	MemoryUtilization uint32 `json:"memoryUtilization"`
	Temperature       uint32 `json:"temperature"`
	MemoryTotal       uint64 `json:"memoryTotal"`
	MemmoryFree       uint64 `json:"memoryFree"`
	MemoryUsed        uint64 `json:"memoryUsed"`
	PowerUsage        uint32 `json:"powerUsage"`
}
