package services

import (
	"fmt"
	"log"
	"testing"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

func TestNVStatusCall(t *testing.T) {
	// 调用 NVStatus 函数

	// 验证 NVML 是否初始化成功
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		t.Fatalf("NVML initialization failed: %v", nvml.ErrorString(ret))
	}
	defer func() {
		ret := nvml.Shutdown()
		if ret != nvml.SUCCESS {
			t.Fatalf("NVML shutdown failed: %v", nvml.ErrorString(ret))
		}
	}()

	// 验证设备数量
	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		t.Fatalf("Unable to get device count: %v", nvml.ErrorString(ret))
	}
	if count <= 0 {
		t.Fatalf("No GPU devices found")
	}
	// 验证每个设备的 UUID
	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			t.Fatalf("Unable to get device at index %d: %v", i, nvml.ErrorString(ret))
		}

		uuid, ret := device.GetUUID()
		if ret != nvml.SUCCESS {
			t.Fatalf("Unable to get uuid of device at index %d: %v", i, nvml.ErrorString(ret))
		}

		if uuid == "" {
			t.Fatalf("Empty UUID for device at index %d", i)
		}
		// 获取 GPU 利用率
		utilization, err := device.GetUtilizationRates()
		if err != nvml.SUCCESS {
			log.Fatalf("Unable to get utilization rates of device at index %d: %v", i, nvml.ErrorString(ret))
		}
		fmt.Println("utilition", utilization)

		// 获取 GPU 温度
		temperature, ret := device.GetTemperature(nvml.TEMPERATURE_GPU)
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to get temperature of device at index %d: %v", i, nvml.ErrorString(ret))
		}
		fmt.Println("temp", temperature)

		// 获取 GPU 内存使用情况
		memory, ret := device.GetMemoryInfo()
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to get memory info of device at index %d: %v", i, nvml.ErrorString(ret))
		}

		fmt.Println("mero", memory)
	}
}
