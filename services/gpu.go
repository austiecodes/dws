package gpu

import (
	"encoding/csv"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/austiecodes/dws/models/types"
)

// GetGPUMetricsService returns the GPU metrics
func GetGPUMetricsService() ([]*types.GPUMetrics, error) {
	cmd := exec.Command("nvidia-smi",
		"--query-gpu=index,name,temperature.gpu,memory.used,memory.total,utilization.gpu,power.draw",
		"--format=csv,noheader,nounits")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute nvidia-smi: %v", err)
	}

	reader := csv.NewReader(strings.NewReader(string(output)))
	reader.TrimLeadingSpace = true
	records, _ := reader.ReadAll()

	var metrics []*types.GPUMetrics
	for _, record := range records {
		cleanedRecord := make([]string, len(record))
		for i, value := range record {
			cleanedRecord[i] = strings.TrimSpace(value)
		}

		gpuID, _ := strconv.ParseUint(cleanedRecord[0], 10, 32)
		name := cleanedRecord[1]
		temperature, _ := strconv.ParseFloat(cleanedRecord[2], 64)
		memoryUsed, _ := strconv.Atoi(cleanedRecord[3])
		memoryTotal, _ := strconv.Atoi(cleanedRecord[4])
		utilization, _ := strconv.ParseFloat(cleanedRecord[5], 64)
		powerDraw, _ := strconv.ParseFloat(cleanedRecord[6], 64)

		metrics = append(metrics, &types.GPUMetrics{
			GPUID:       uint(gpuID),
			Name:        name,
			Temperature: temperature,
			MemoryUsed:  memoryUsed,
			MemoryTotal: memoryTotal,
			Utilization: utilization,
			PowerDraw:   powerDraw,
		})
	}

	return metrics, nil
}
