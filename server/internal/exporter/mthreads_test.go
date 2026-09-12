package exporter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/prometheus/client_golang/prometheus"
	pb "vgpu/api/v1"
	"vgpu/internal/biz"
)

func mthreadsDevice() *biz.DeviceInfo {
	return &biz.DeviceInfo{Id: "mtt-node-mthreads-1", AliasId: "mtt-node-mthreads-1", Index: 1,
		NodeName: "mtt-node", Type: "MTT S4000", Provider: biz.MthreadsGPUDevice, Devmem: 49152, Devcore: 100, Count: 100}
}

func TestMthreadsPhysicalQueriesUseNativeOrdinalAndPreserveIdleZero(t *testing.T) {
	for _, present := range []bool{false, true} {
		t.Run(fmt.Sprintf("present=%t", present), func(t *testing.T) {
			device := mthreadsDevice()
			responses := map[string]*pb.InstantResponse{}
			if present {
				for field, value := range map[string]float32{"GPU_UTIL": 0, "GPU_TEMP": 35, "MEMORY_TEMP": 38, "POWER_USAGE": 30, "FB_USED": 0, "FB_TOTAL": 49152} {
					responses[fmt.Sprintf(`avg(DCGM_FI_DEV_%s{Hostname="mtt-node",gpu="1",device="mtgpu1"})`, field)] = instantValue(value)
				}
			}
			generator := newDeviceMetricsTestGenerator(device, responses)
			t.Cleanup(func() { deleteTrackedTestCells(generator) })
			if err := generator.GenerateDeviceMetrics(context.Background()); err != nil {
				t.Fatal(err)
			}
			labels := []string{device.NodeName, device.Provider, device.Type, device.Id, "", "mtgpu1"}
			assertTrackedGaugeValue(t, generator, HamiVmemorySize, labels, 49152)
			for gauge, value := range map[*prometheus.GaugeVec]float64{HamiMemoryUsed: 0, HamiMemorySize: 49152, HamiMemoryUtil: 0, HamiCoreUtil: 0, HamiDeviceTemperature: 35, HamiDeviceMemoryTemperature: 38, HamiDevicePower: 30} {
				assertGaugeTracked(t, generator, gauge, labels, present)
				if present {
					assertTrackedGaugeValue(t, generator, gauge, labels, value)
				}
			}
		})
	}
}

func TestMthreadsContainerAllocationDoesNotBorrowCardUsageOrPrefixMatchAnotherCard(t *testing.T) {
	device := mthreadsDevice()
	logger := log.NewStdLogger(io.Discard)
	querier := &fakeInstantQuerier{responsesByQuery: map[string]*pb.InstantResponse{}}
	generator := &MetricsGenerator{
		nodeUsecase: biz.NewNodeUsecase(&fakeNodeRepo{devices: []*biz.DeviceInfo{device}}, logger),
		podUsecase: biz.NewPodUseCase(&fakePodRepo{containers: []*biz.Container{{Name: "worker", PodName: "train", PodUID: "uid", Namespace: "research",
			ContainerDevices: biz.ContainerDevices{
				{UUID: device.AliasId, Type: "Mthreads", Usedmem: 4096, Usedcores: 25},
				{UUID: "mtt-node-mthreads-10", Type: "Mthreads", Usedmem: 8192, Usedcores: 50},
			}}}}, logger), monitorService: querier, log: log.NewHelper(logger),
	}
	t.Cleanup(func() { deleteTrackedTestCells(generator) })
	if err := generator.GenerateContainerMetrics(context.Background()); err != nil {
		t.Fatal(err)
	}
	labels := []string{device.NodeName, device.Provider, device.Type, device.Id, "train", "worker", "research"}
	allocatedLabels := append(append([]string{}, labels...), "worker:uid")
	assertTrackedGaugeValue(t, generator, HamiContainerVgpuAllocated, allocatedLabels, 1)
	assertTrackedGaugeValue(t, generator, HamiContainerVmemoryAllocated, allocatedLabels, 4096)
	assertTrackedGaugeValue(t, generator, HamiContainerVcoreAllocated, allocatedLabels, 25)
	for _, gauge := range []*prometheus.GaugeVec{HamiContainerCoreUsed, HamiContainerCoreUtil, HamiContainerMemoryUsed, HamiContainerMemoryUtil} {
		assertGaugeTracked(t, generator, gauge, labels, false)
	}
	if len(querier.queries) != 0 {
		t.Fatalf("unsupported workload telemetry queried physical data: %v", querier.queries)
	}
	if _, err := generator.taskCoreUsed(context.Background(), biz.MthreadsGPUDevice, "research", "train", "worker", "uid", device.Id, device.NodeName, 1); !errors.Is(err, errWorkloadTelemetryUnsupported) {
		t.Fatal(err)
	}
}

func TestMthreadsCollectionFailureRemainsAnError(t *testing.T) {
	device := mthreadsDevice()
	g := newDeviceMetricsTestGenerator(device, nil)
	g.monitorService = &fakeInstantQuerier{query: func(context.Context, *pb.QueryInstantRequest) (*pb.InstantResponse, error) {
		return nil, errors.New("collector failed")
	}}
	t.Cleanup(func() { deleteTrackedTestCells(g) })
	if err := g.GenerateDeviceMetrics(context.Background()); err == nil {
		t.Fatal("collection failure reported success")
	}
	labels := []string{device.NodeName, device.Provider, device.Type, device.Id, "", "mtgpu1"}
	assertGaugeTracked(t, g, HamiCoreUtil, labels, false)
	assertTrackedGaugeValue(t, g, HamiVgpuCount, labels, 100)
}
