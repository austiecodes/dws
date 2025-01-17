package services

import (
	"log"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/austiecodes/dws/models/types"
	"github.com/gin-gonic/gin"
)

// // GetGPUMetricsService returns the GPU metrics
// func GetGPUMetricsService() ([]*types.GPUMetrics, error) {
// 	cmd := exec.Command("nvidia-smi",
// 		"--query-gpu=index,name,temperature.gpu,memory.used,memory.total,utilization.gpu,power.draw",
// 		"--format=csv,noheader,nounits")

// 	output, err := cmd.Output()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to execute nvidia-smi: %v", err)
// 	}

// 	reader := csv.NewReader(strings.NewReader(string(output)))
// 	reader.TrimLeadingSpace = true
// 	records, _ := reader.ReadAll()

// 	var metrics []*types.GPUMetrics
// 	for _, record := range records {
// 		cleanedRecord := make([]string, len(record))
// 		for i, value := range record {
// 			cleanedRecord[i] = strings.TrimSpace(value)
// 		}

// 		gpuID, _ := strconv.ParseUint(cleanedRecord[0], 10, 32)
// 		name := cleanedRecord[1]
// 		temperature, _ := strconv.ParseFloat(cleanedRecord[2], 64)
// 		memoryUsed, _ := strconv.Atoi(cleanedRecord[3])
// 		memoryTotal, _ := strconv.Atoi(cleanedRecord[4])
// 		utilization, _ := strconv.ParseFloat(cleanedRecord[5], 64)
// 		powerDraw, _ := strconv.ParseFloat(cleanedRecord[6], 64)

// 		metrics = append(metrics, &types.GPUMetrics{
// 			GPUID:       uint(gpuID),
// 			Name:        name,
// 			Temperature: temperature,
// 			MemoryUsed:  memoryUsed,
// 			MemoryTotal: memoryTotal,
// 			Utilization: utilization,
// 			PowerDraw:   powerDraw,
// 		})
// 	}

// 	return metrics, nil
// }

func NVStatus(c *gin.Context) ([]*types.GPUMetrics, error) {
	ret := make([]*types.GPUMetrics, 0)
	err := nvml.Init()
	if err != nvml.SUCCESS {
		log.Fatalf("Unable to initialize NVML: %v", nvml.ErrorString(err))
	}
	defer func() {
		// Use a different variable name to avoid shadowing
		err := nvml.Shutdown()
		if err != nvml.SUCCESS {
			log.Fatalf("Unable to shutdown NVML: %v", nvml.ErrorString(err))
		}
	}()
	count, err := nvml.DeviceGetCount()
	if err != nvml.SUCCESS {
		log.Fatalf("Unable to get device count: %v", nvml.ErrorString(err))
	}

	for i := 0; i < count; i++ {
		var gpu types.GPUMetrics
		device, err := nvml.DeviceGetHandleByIndex(i)
		if err != nvml.SUCCESS {
			log.Fatalf("Unable to get device at index %d: %v", i, nvml.ErrorString(err))
		}
		gpu.ID = i
		gpu.Name, _ = device.GetName()
		utilization, _ := device.GetUtilizationRates()
		gpu.GPUUtilization = utilization.Gpu
		gpu.MemoryUtilization = utilization.Memory
		gpu.Temperature, _ = device.GetTemperature(nvml.TEMPERATURE_GPU)
		gpu.PowerUsage, _ = device.GetPowerUsage()

		ret = append(ret, &gpu)
	}

	return ret, nil
}
