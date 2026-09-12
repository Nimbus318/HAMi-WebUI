package util

import (
	"io"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMLUProfileUsesSchedulerSlotAnd256MiBUnitsBeforeRuntimeCompletion(t *testing.T) {
	setSupportDeviceForTest(t, CambriconGPUDevice, "CAMBRICON_DSMLU_PROFILE")
	for _, instance := range []string{"", "43_1794_2_7"} {
		// Cambricon DP 5731364a23f4938fbe42c150b51de3f16eaf605d emits
		// profileID_handle_slot_instanceID, handle=(instanceID<<8)|slot.
		pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
			AssignedNodeAnnotations: "mlu-node", "CAMBRICON_DSMLU_PROFILE": "2_25_16",
			DsmluProfileAndInstance: instance,
		}}, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "worker"}}}}
		devices, err := DecodePodDevices(pod, log.NewHelper(log.NewStdLogger(io.Discard)))
		if err != nil {
			t.Fatal(err)
		}
		got := devices[CambriconGPUDevice]
		if len(got) != 1 || len(got[0]) != 1 {
			t.Fatalf("allocation=%#v", devices)
		}
		allocation := got[0][0]
		if allocation.UUID != "mlu-node-cambricon-mlu-2" || allocation.Usedmem != 4096 || allocation.Usedcores != 25 {
			t.Fatalf("allocation=%#v", allocation)
		}
	}
}

func TestMLUMalformedProfileDoesNotPanicOrOverflow(t *testing.T) {
	for _, profile := range []string{"2_25_16_", "2_25", "2_bad_16", "-1_25_16", "2_101_16", "2_25_8388608"} {
		if _, err := DecodeMLUContainerDevices(profile, "node"); err == nil {
			t.Fatalf("accepted %q", profile)
		}
	}
	zero, err := DecodeMLUContainerDevices("2_0_16", "node")
	if err != nil || zero[0].Usedcores != 0 {
		t.Fatalf("zero reserved cores changed: %#v %v", zero, err)
	}
}
