package managers

import "github.com/NVIDIA/go-nvml/pkg/nvml"

type GPUManager struct {
	Devices []*nvml.Device // 存储所有 GPU 设备
}
