package services

import (
	"log"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/austiecodes/dws/models/types"
	"github.com/gin-gonic/gin"
)

func GPUStatus(c *gin.Context) ([]*types.GPUMetrics, error) {
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
		memory, _ := device.GetMemoryInfo()
		gpu.MemoryTotal = memory.Total
		gpu.MemmoryFree = memory.Free
		gpu.MemoryUsed = memory.Used

		ret = append(ret, &gpu)
	}

	return ret, nil
}
