package mthreads

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/prometheus/common/model"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type fakeQuerier struct {
	value model.Value
	err   error
	query string
}

func (q *fakeQuerier) Query(_ context.Context, query string) (model.Value, error) {
	q.query = query
	return q.value, q.err
}

func testNode() *corev1.Node {
	return &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "mtt-node"}, Status: corev1.NodeStatus{Capacity: corev1.ResourceList{
		CoreResource: resource.MustParse("32"), MemoryResource: resource.MustParse("192"),
	}}}
}

func TestCapacityInventoryPreservesSchedulerIDsWithoutTelemetry(t *testing.T) {
	for _, failure := range []error{nil, errors.New("collector unavailable")} {
		provider := NewMthreads(&fakeQuerier{err: failure}, log.NewHelper(log.NewStdLogger(io.Discard)), "")
		devices, err := provider.FetchDevices(testNode())
		if err != nil || len(devices) != 2 {
			t.Fatalf("devices=%#v error=%v", devices, err)
		}
		for index, device := range devices {
			if device.Index != uint(index) || device.Count != 100 || device.Devmem != 49152 || device.Devcore != 100 || device.AliasId != device.ID {
				t.Fatalf("unexpected inventory: %#v", device)
			}
		}
		if devices[1].ID != "mtt-node-mthreads-1" {
			t.Fatal(devices[1].ID)
		}
	}
}

func TestNativePhysicalLabelsOnlyEnrichDisplayMetadata(t *testing.T) {
	query := &fakeQuerier{value: model.Vector{&model.Sample{Metric: model.Metric{
		"Hostname": "mtt-node", "gpu": "1", "device": "mtgpu1", "UUID": "unrelated-exporter-uuid",
		"modelName": "MTT S4000", "DCGM_FI_DRIVER_VERSION": "synthetic",
	}}}}
	provider := NewMthreads(query, log.NewHelper(log.NewStdLogger(io.Discard)), "")
	devices, err := provider.FetchDevices(testNode())
	if err != nil {
		t.Fatal(err)
	}
	if query.query != `DCGM_FI_DEV_GPU_UTIL{Hostname="mtt-node",device=~"mtgpu[0-9]+"}` {
		t.Fatal(query.query)
	}
	if devices[1].Type != "MTT S4000" || devices[1].Driver != "synthetic" || devices[1].ID != "mtt-node-mthreads-1" {
		t.Fatalf("device=%#v", devices[1])
	}
	if devices[0].Type != "Mthreads" {
		t.Fatalf("unobserved card gained guessed model: %#v", devices[0])
	}
}

func TestUnsupportedWholeCardAndFractionalTopologyAreNotInferred(t *testing.T) {
	provider := NewMthreads(nil, log.NewHelper(log.NewStdLogger(io.Discard)), "")
	node := testNode()
	node.Status.Capacity = corev1.ResourceList{"mthreads.com/gpu": resource.MustParse("4")}
	devices, err := provider.FetchDevices(node)
	if err != nil || len(devices) != 0 {
		t.Fatalf("whole-card topology inferred: %#v %v", devices, err)
	}
	node = testNode()
	node.Status.Capacity[CoreResource] = resource.MustParse("17")
	if _, err := provider.FetchDevices(node); err == nil {
		t.Fatal("partial card topology accepted")
	}
}
