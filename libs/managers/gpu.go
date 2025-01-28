package managers

import "github.com/NVIDIA/go-nvml/pkg/nvml"

type GPUManager struct {
	Devices []*nvml.Device //store all gpu devices
}
