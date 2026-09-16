package nvidia

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
)

// HAMi's scheduler records the MIG device it reserved for each allocated GPU
// (pkg/device/nvidia/mig_allocations.go); the device plugin fills in migUUID.
const MigAllocationsAnnotation = "hami.io/vgpu-mig-allocations"

type migAllocation struct {
	ContainerIndex int       `json:"containerIndex"`
	DeviceIndex    int       `json:"deviceIndex"`
	GPUUUID        string    `json:"gpuUUID"`
	Profile        string    `json:"profile"`
	Placement      placement `json:"placement"`
}

type placement struct {
	Start int32 `json:"start"`
	Size  int32 `json:"size"`
}

// MigDevice is one MIG device HAMi reserved on a GPU.
type MigDevice struct {
	Profile string
	Start   int32
	Size    int32
}

// MigDevices returns the MIG device reserved for each allocated GPU, by
// container slot and device index within that slot.
func MigDevices(pod *corev1.Pod) map[int]map[int]MigDevice {
	value := pod.Annotations[MigAllocationsAnnotation]
	var allocations []migAllocation
	if value == "" || json.Unmarshal([]byte(value), &allocations) != nil {
		return nil
	}
	devices := map[int]map[int]MigDevice{}
	for _, allocation := range allocations {
		if allocation.Profile == "" || allocation.ContainerIndex < 0 || allocation.DeviceIndex < 0 {
			continue
		}
		if devices[allocation.ContainerIndex] == nil {
			devices[allocation.ContainerIndex] = map[int]MigDevice{}
		}
		devices[allocation.ContainerIndex][allocation.DeviceIndex] = MigDevice{
			Profile: allocation.Profile,
			Start:   allocation.Placement.Start,
			Size:    allocation.Placement.Size,
		}
	}
	return devices
}
