package util

import (
	"io"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMthreadsAllocationKeepsZeroAndNormalizesSixteenUnitScale(t *testing.T) {
	for encoded, want := range map[string]int32{"0": 0, "4": 25, "8": 50, "16": 100} {
		devices, err := DecodeMthreadsContainerDevices("node-mthreads-1,Mthreads,4096,"+encoded+":", "1")
		if err != nil || len(devices) != 1 {
			t.Fatalf("decoded=%v error=%v", devices, err)
		}
		if devices[0].Usedcores != want || devices[0].Usedmem != 4096 || devices[0].UUID != "node-mthreads-1" {
			t.Fatalf("producer units lost: %#v", devices[0])
		}
	}
	for _, encoded := range []string{"node,Mthreads,4096,-1:", "node,Mthreads,4096,17:", "node,Mthreads,NaN,4:", "node,Other,4096,4:"} {
		if _, err := DecodeMthreadsContainerDevices(encoded, ""); err == nil {
			t.Fatalf("accepted %q", encoded)
		}
	}
}

func TestMthreadsPodDecoderRetainsContainerSlotsAndIgnoresPendingAssignment(t *testing.T) {
	const key = "hami.io/mthreads-vgpu-devices-allocated"
	setSupportDeviceForTest(t, MthreadsGPUDevice, key)
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{key: ";node-mthreads-0,Mthreads,4096,4:;node-mthreads-1,Mthreads,8192,8:;"}},
		Spec: corev1.PodSpec{InitContainers: []corev1.Container{{Name: "init"}}, Containers: []corev1.Container{{Name: "a"}, {Name: "b"}}}}
	logger := log.NewHelper(log.NewStdLogger(io.Discard))
	devices, err := DecodePodDevices(pod, logger)
	if err != nil {
		t.Fatal(err)
	}
	got := devices[MthreadsGPUDevice]
	if len(got) != 3 || len(got[0]) != 0 || got[1][0].Usedcores != 25 || got[2][0].Usedmem != 8192 {
		t.Fatalf("container mapping: %#v", got)
	}
	pod.Annotations["hami.io/mthreads-vgpu-devices-to-allocate"] = pod.Annotations[key]
	delete(pod.Annotations, key)
	devices, err = DecodePodDevices(pod, logger)
	if err != nil || len(devices) != 0 {
		t.Fatalf("pending allocation became delivered: %#v %v", devices, err)
	}
}
