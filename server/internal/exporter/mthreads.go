// Copyright 2026 The HAMi Authors.
// SPDX-License-Identifier: Apache-2.0
package exporter

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"vgpu/internal/biz"
)

// MT DCGM's official profile uses Hostname + gpu + mtgpu<N>. Its UUID is not
// HAMi's capacity-derived node-mthreads-N ID. All series here stay device-scoped.
func (s *MetricsGenerator) generateMthreadsDeviceMetrics(ctx context.Context, device *biz.DeviceInfo, driver, deviceNo string) {
	labels := []string{device.NodeName, biz.MthreadsGPUDevice, device.Type, device.Id, driver, deviceNo}
	read := func(field string) (float32, bool) {
		query := fmt.Sprintf(`avg(DCGM_FI_DEV_%s{Hostname=%q,gpu="%d",device=%q})`, field, device.NodeName, device.Index, deviceNo)
		value, err := s.queryRequiredInstantVal(ctx, query)
		return value, err == nil
	}
	for field, gauges := range map[string][]*prometheus.GaugeVec{
		"GPU_TEMP": {HamiDeviceTemperature}, "MEMORY_TEMP": {HamiDeviceMemoryTemperature},
		"POWER_USAGE": {HamiDevicePower},
		"GPU_UTIL":    {HamiCoreUsed, HamiCoreUtil, HamiCoreUsedAvg, HamiCoreUtilAvg},
	} {
		if value, present := read(field); present {
			for _, gauge := range gauges {
				s.set(gauge, float64(value), labels...)
			}
		}
	}
	used, usedPresent := read("FB_USED")
	total, totalPresent := read("FB_TOTAL")
	if usedPresent {
		s.set(HamiMemoryUsed, float64(used), labels...)
	}
	if totalPresent && total > 0 {
		s.set(HamiMemorySize, float64(total), labels...)
		s.set(HamiVMemoryScaling, roundToOneDecimal(float64(device.Devmem)/float64(total)), labels...)
		if usedPresent {
			s.set(HamiMemoryUtil, roundToOneDecimal(100*float64(used)/float64(total)), labels...)
		}
	}
}
