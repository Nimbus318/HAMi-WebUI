package mlu

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"vgpu/internal/data/prom"
)

// HAMi 88b118e565a9effaa27f81669da291c5508fc5e4, pkg/device/cambricon/device.go:
// GetNodeDevices Count=100 and MemoryFactor=256, independently of exporter bytes.
func TestSMLUInventoryMatchesHAMiSlotsAndNativeMemoryUnits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Error(err)
		}
		if r.Form.Get("query") != `mlu_health{node="mlu-node"}` {
			t.Errorf("query=%q", r.Form.Get("query"))
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(w, `{"status":"success","data":{"resultType":"vector","result":[{"metric":{"mlu":"0","uuid":"MLU-physical-0","model":"MLU370-X8","driver":"synthetic","unhealth_reason":""},"value":[0,"1"]},{"metric":{"mlu":"1","uuid":"MLU-physical-1","model":"MLU370-X8","driver":"synthetic","unhealth_reason":""},"value":[0,"1"]}]}}`); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()
	logger := log.NewStdLogger(io.Discard)
	client, err := prom.NewClient(server.URL, time.Second, prom.HTTPConfig{}, logger)
	if err != nil {
		t.Fatal(err)
	}
	node := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "mlu-node"}, Status: corev1.NodeStatus{Capacity: corev1.ResourceList{
		CambriconDeviceCoreAnnos: resource.MustParse("200"), CambriconDeviceMemAnnos: resource.MustParse("384"),
	}}}
	devices, err := NewCambricon(client, log.NewHelper(logger), "").FetchDevices(node)
	if err != nil || len(devices) != 2 {
		t.Fatalf("devices=%#v err=%v", devices, err)
	}
	for _, device := range devices {
		if device.Count != 100 || device.Devmem != 49152 || device.Devcore != 100 {
			t.Fatalf("HAMi per-card inventory disagrees: %#v", device)
		}
	}
	if devices[1].AliasId != "mlu-node-cambricon-mlu-1" || devices[1].ID != "MLU-physical-1" {
		t.Fatalf("identities lost: %#v", devices[1])
	}
}
