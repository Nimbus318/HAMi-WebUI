package service

import (
	"context"
	"os"
	"testing"

	pb "vgpu/api/v1"
	"vgpu/internal/biz"
	"vgpu/internal/devicecatalog"

	"github.com/go-kratos/kratos/v2/log"
)

func TestContainerRepliesDescribeInterpretedAllocations(t *testing.T) {
	containers := []*biz.Container{
		{Name: "template", PodUID: "pod-1", Status: biz.ContainerStatusSuccess, ContainerDevices: biz.ContainerDevices{
			{UUID: "B3-0", Type: "Ascend910B3", Usedmem: 16384, Usedcores: 25, Vendor: biz.AscendGPUDevice, Shape: "template", Template: "vir05_1c_16g"},
		}},
		{Name: "unknown", PodUID: "pod-2", Status: biz.ContainerStatusSuccess, ContainerDevices: biz.ContainerDevices{
			{UUID: "B3-1", Type: "Ascend910B3", Usedmem: 32768, Vendor: biz.AscendGPUDevice, Shape: "template", Template: "vir10_3c_32g", CoreAllocationUnknown: true, CoreReason: "catalog_unavailable"},
			{UUID: "B3-2", Type: "Ascend910B3", Usedmem: 65536, Usedcores: 100, Vendor: biz.AscendGPUDevice, Shape: "whole"},
		}},
	}
	service := NewContainerService(
		biz.NewNodeUsecase(&containerTestNodeRepo{}, log.DefaultLogger),
		biz.NewPodUseCase(&containerTestPodRepo{containers: containers}, log.DefaultLogger),
	)
	reply, err := service.GetAllContainers(context.Background(), &pb.GetAllContainersReq{})
	if err != nil || len(reply.Items) != 2 {
		t.Fatalf("GetAllContainers = %v, %v", reply, err)
	}
	byName := map[string]*pb.ContainerReply{}
	for _, item := range reply.Items {
		byName[item.Name] = item
	}
	if item := byName["template"]; item.AllocationShape != "template" || item.Template != "vir05_1c_16g" || item.AllocatedCoresReason != "" || !item.GetAllocatedCoresKnown() {
		t.Fatalf("template container = %+v", item)
	}
	if item := byName["unknown"]; item.AllocationShape != "" || item.Template != "" || item.AllocatedCoresReason != "catalog_unavailable" || item.GetAllocatedCoresKnown() {
		t.Fatalf("mixed container = %+v", item)
	}
}

func TestUnconfiguredDevicesAddNoSchedulableCapacity(t *testing.T) {
	configured := &biz.DeviceInfo{Id: "B3-0", AliasId: "B3-0", Type: "Ascend910B3", Count: 4, Devmem: 65536, NodeName: "node-a", NodeUid: "node-uid-a"}
	stale := &biz.DeviceInfo{Id: "B3-9", AliasId: "B3-9", Type: "Ascend910B9", Count: 4, Devmem: 65536, NodeName: "node-a", NodeUid: "node-uid-a", Unconfigured: true}
	nodeRepo := &capacityTestNodeRepo{
		nodes:   []*biz.Node{{Name: "node-a", Uid: "node-uid-a", Devices: []*biz.DeviceInfo{configured, stale}}},
		devices: []*biz.DeviceInfo{configured, stale},
	}
	podRepo := &capacityTestPodRepo{}
	nodeUsecase := biz.NewNodeUsecase(nodeRepo, log.DefaultLogger)
	podUsecase := biz.NewPodUseCase(podRepo, log.DefaultLogger)
	nodes := NewNodeService(nodeUsecase, podUsecase, biz.NewSummaryUseCase(nodeRepo, podRepo, log.DefaultLogger))

	node, err := nodes.GetNode(context.Background(), &pb.GetNodeReq{Uid: "node-uid-a"})
	if err != nil || node.VgpuTotal != 4 || node.MemoryTotal != 65536 {
		t.Fatalf("node totals = %+v, %v", node, err)
	}
	summary, err := nodes.GetSummary(context.Background(), &pb.GetSummaryReq{})
	if err != nil || summary.VgpuTotal != 4 || summary.MemoryTotal != 65536 {
		t.Fatalf("summary = %+v, %v", summary, err)
	}
	cards, err := NewCardService(nodeUsecase, podUsecase).GetAllGPUs(context.Background(), &pb.GetAllGpusReq{})
	if err != nil || len(cards.List) != 2 || cards.List[0].Unconfigured || !cards.List[1].Unconfigured {
		t.Fatalf("cards = %+v, %v", cards, err)
	}
}

func TestDeviceConfigReplyListsModelsAndState(t *testing.T) {
	content, err := os.ReadFile("../devicecatalog/testdata/hami-v2.10.0.yaml")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := devicecatalog.Parse(content)
	if err != nil {
		t.Fatal(err)
	}
	ref := devicecatalog.Ref{Namespace: "kube-system", Name: "hami-scheduler-device", Key: "device-config.yaml"}
	reply, err := NewDeviceConfigService(devicecatalog.Static{Current: &devicecatalog.Snapshot{State: devicecatalog.StateLoaded, Ref: ref, ResourceVersion: "42", Ascend: parsed.Ascend}}).
		GetDeviceConfig(context.Background(), &pb.GetDeviceConfigReq{})
	if err != nil || reply.State != "loaded" || reply.Namespace != "kube-system" || reply.ResourceVersion != "42" || len(reply.AscendModels) != 7 {
		t.Fatalf("reply = %+v, %v", reply, err)
	}
	var b3 *pb.AscendModel
	for _, model := range reply.AscendModels {
		if model.CommonWord == "Ascend910B3" {
			b3 = model
		}
	}
	if b3 == nil || b3.AiCore != 20 || len(b3.Templates) != 2 || b3.Templates[0].GetComputeShare() != 25 || b3.Templates[1].Name != "vir10_3c_32g" {
		t.Fatalf("910B3 = %+v", b3)
	}

	forbidden, _ := NewDeviceConfigService(devicecatalog.Static{Current: &devicecatalog.Snapshot{State: devicecatalog.StateForbidden, Reason: "no RBAC", Ref: ref}}).
		GetDeviceConfig(context.Background(), &pb.GetDeviceConfigReq{})
	if forbidden.State != "forbidden" || forbidden.Reason != "no RBAC" || len(forbidden.AscendModels) != 0 {
		t.Fatalf("forbidden reply = %+v", forbidden)
	}
}
