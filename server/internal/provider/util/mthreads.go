// Copyright 2026 The HAMi Authors.
// SPDX-License-Identifier: Apache-2.0
package util

import (
	"fmt"
	"strconv"
	"strings"
)

// DecodeMthreadsContainerDevices normalizes only the UI percentage scale.
// Memory is already MiB in retained HAMi annotations. Zero cores remain zero:
// HAMi records no reserved core weight, not an exclusive whole-card allocation.
func DecodeMthreadsContainerDevices(encoded, priority string) (ContainerDevices, error) {
	var devices ContainerDevices
	for index, segment := range strings.Split(encoded, OneContainerMultiDeviceSplitSymbol) {
		if segment == "" {
			continue
		}
		fields := strings.Split(segment, ",")
		if len(fields) != 4 || fields[0] == "" || fields[1] != MthreadsGPUDevice {
			return nil, fmt.Errorf("invalid Mthreads allocation segment %q", segment)
		}
		memory, memoryErr := strconv.ParseInt(fields[2], 10, 32)
		cores, coreErr := strconv.ParseInt(fields[3], 10, 32)
		if memoryErr != nil || coreErr != nil || memory < 0 || cores < 0 || cores > 16 {
			return nil, fmt.Errorf("invalid Mthreads allocation quantities in %q", segment)
		}
		devices = append(devices, ContainerDevice{
			Idx: index, UUID: fields[0], Type: MthreadsGPUDevice, Usedmem: int32(memory),
			Usedcores: int32(cores * 100 / 16), Priority: priority,
		})
	}
	return devices, nil
}
