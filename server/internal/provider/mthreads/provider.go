// Copyright 2026 The HAMi Authors.
// SPDX-License-Identifier: Apache-2.0
package mthreads

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/prometheus/common/model"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"vgpu/internal/biz"
	"vgpu/internal/provider/util"
)

const (
	CoreResource   = "mthreads.com/sgpu-core"
	MemoryResource = "mthreads.com/sgpu-memory"
	CoresPerCard   = 16
	MemoryUnitMiB  = 512
)

func init() {
	util.InRequestDevices[util.MthreadsGPUDevice] = "hami.io/mthreads-vgpu-devices-to-allocate"
	util.SupportDevices[util.MthreadsGPUDevice] = "hami.io/mthreads-vgpu-devices-allocated"
}

type instantQuerier interface {
	Query(context.Context, string) (model.Value, error)
}

type Mthreads struct {
	prom     instantQuerier
	log      *log.Helper
	selector string
}

func NewMthreads(prom instantQuerier, logger *log.Helper, selector string) *Mthreads {
	if selector == "" {
		selector = "mthreads.com/gpu-node=true"
	}
	return &Mthreads{prom: prom, log: logger, selector: selector}
}

func (m *Mthreads) GetNodeDevicePluginLabels() (labels.Selector, error) {
	return labels.Parse(m.selector)
}

func (m *Mthreads) GetProvider() string { return biz.MthreadsGPUDevice }

// FetchDevices follows HAMi's capacity-based inventory for homogeneous sGPU
// nodes. It does not infer vendor whole-card allocations or remap card ordinals
// from mixed-mode labels. See docs/providers/mthreads-experimental.md.
func (m *Mthreads) FetchDevices(node *corev1.Node) ([]*util.DeviceInfo, error) {
	coresQuantity := node.Status.Capacity[CoreResource]
	cores, ok := coresQuantity.AsInt64()
	if !ok || cores <= 0 {
		return nil, nil
	}
	memoryQuantity := node.Status.Capacity[MemoryResource]
	memoryUnits, ok := memoryQuantity.AsInt64()
	if !ok || memoryUnits <= 0 || memoryUnits > math.MaxInt64/MemoryUnitMiB || cores%CoresPerCard != 0 {
		return nil, fmt.Errorf("node %s has no supported homogeneous Mthreads sGPU inventory", node.Name)
	}
	count := cores / CoresPerCard
	memoryMiB := memoryUnits * MemoryUnitMiB / count
	if memoryMiB > math.MaxInt32 || memoryUnits*MemoryUnitMiB%count != 0 {
		return nil, fmt.Errorf("node %s Mthreads per-card memory is not representable", node.Name)
	}
	devices := make([]*util.DeviceInfo, 0, count)
	for i := int64(0); i < count; i++ {
		id := fmt.Sprintf("%s-mthreads-%d", node.Name, i)
		devices = append(devices, &util.DeviceInfo{
			ID: id, AliasId: id, Index: uint(i), Count: 100, Devmem: int32(memoryMiB),
			Devcore: 100, Type: util.MthreadsGPUDevice, Mode: "sgpu", Health: true,
		})
	}
	// Telemetry supplies display metadata only. Missing telemetry never removes
	// scheduler inventory, changes the scheduler identity or claims card health.
	if m.prom != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		query := fmt.Sprintf(`DCGM_FI_DEV_GPU_UTIL{Hostname=%q,device=~"mtgpu[0-9]+"}`, node.Name)
		value, err := m.prom.Query(ctx, query)
		if err != nil {
			m.log.Warnf("Mthreads display metadata unavailable on %s: %v", node.Name, err)
		} else if vector, ok := value.(model.Vector); ok {
			for _, sample := range vector {
				index, err := strconv.ParseUint(string(sample.Metric["gpu"]), 10, 32)
				if err != nil || index >= uint64(len(devices)) || string(sample.Metric["Hostname"]) != node.Name ||
					string(sample.Metric["device"]) != fmt.Sprintf("mtgpu%d", index) {
					continue
				}
				if name := string(sample.Metric["modelName"]); name != "" {
					devices[index].Type = name
				}
				devices[index].Driver = string(sample.Metric["DCGM_FI_DRIVER_VERSION"])
			}
		}
	}
	return devices, nil
}
