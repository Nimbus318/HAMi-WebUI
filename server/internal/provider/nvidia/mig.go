package nvidia

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
)

// HAMi's scheduler records the MIG device it reserved for each allocated GPU
// (pkg/device/nvidia/mig_allocations.go); the device plugin fills in migUUID.
const MigAllocationsAnnotation = "hami.io/vgpu-mig-allocations"

type migAllocation struct {
	ContainerIndex int    `json:"containerIndex"`
	DeviceIndex    int    `json:"deviceIndex"`
	GPUUUID        string `json:"gpuUUID"`
	Profile        string `json:"profile"`
}

// MigProfiles returns the reserved MIG profile of each allocated GPU, by
// container slot and device index within that slot.
func MigProfiles(pod *corev1.Pod) map[int]map[int]string {
	value := pod.Annotations[MigAllocationsAnnotation]
	var allocations []migAllocation
	if value == "" || json.Unmarshal([]byte(value), &allocations) != nil {
		return nil
	}
	profiles := map[int]map[int]string{}
	for _, allocation := range allocations {
		if allocation.Profile == "" || allocation.ContainerIndex < 0 || allocation.DeviceIndex < 0 {
			continue
		}
		if profiles[allocation.ContainerIndex] == nil {
			profiles[allocation.ContainerIndex] = map[int]string{}
		}
		profiles[allocation.ContainerIndex][allocation.DeviceIndex] = allocation.Profile
	}
	return profiles
}
