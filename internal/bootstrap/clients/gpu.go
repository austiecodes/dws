package clients

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/austiecodes/dws/lib/managers"
	"github.com/austiecodes/dws/lib/resources"
)

type GPUConfig struct {
	Enabled bool `toml:"enabled"`
}

type GPUClient struct {
	config GPUConfig
}

func NewGPUClient() *GPUClient {
	return &GPUClient{}
}

func (g *GPUClient) LoadConfig() error {
	var config struct {
		GPU GPUConfig `toml:"gpu"`
	}
	if _, err := toml.DecodeFile("conf/gpu.toml", &config); err != nil {
		return fmt.Errorf("error loading GPU config: %w", err)
	}
	g.config = config.GPU
	return nil
}

func (g *GPUClient) Init() error {
	if !g.config.Enabled {
		return nil
	}

	if ret := nvml.Init(); ret != nvml.SUCCESS {
		return fmt.Errorf("failed to initialize NVML: %v", nvml.ErrorString(ret))
	}

	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return fmt.Errorf("failed to get GPU device count: %v", nvml.ErrorString(ret))
	}

	resources.GPUManager = &managers.GPUManager{
		Devices: make([]*nvml.Device, 0, count),
	}

	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			return fmt.Errorf("failed to get GPU device at index %d: %v", i, nvml.ErrorString(ret))
		}
		resources.GPUManager.Devices = append(resources.GPUManager.Devices, &device)
	}

	return nil
}
