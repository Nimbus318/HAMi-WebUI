package nvidia

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMigProfilesFollowHAMisReservations(t *testing.T) {
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
		MigAllocationsAnnotation: `[{"containerIndex":0,"deviceIndex":0,"gpuUUID":"GPU-0","profile":"1g.10gb","placement":{"start":0,"size":1},"migUUID":"MIG-0"},
			{"containerIndex":1,"deviceIndex":1,"gpuUUID":"GPU-1","profile":"3g.40gb","placement":{"start":0,"size":4}},
			{"containerIndex":1,"deviceIndex":0,"gpuUUID":"GPU-0","profile":""}]`,
	}}}
	profiles := MigProfiles(pod)
	if profiles[0][0] != "1g.10gb" || profiles[1][1] != "3g.40gb" {
		t.Fatalf("profiles = %v", profiles)
	}
	if _, ok := profiles[1][0]; ok {
		t.Fatalf("an entry without a profile was kept: %v", profiles)
	}
	for _, value := range []string{"", "not json", `[{"containerIndex":-1,"profile":"1g.10gb"}]`} {
		got := MigProfiles(&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{MigAllocationsAnnotation: value}}})
		if len(got) != 0 {
			t.Fatalf("%q gave %v", value, got)
		}
	}
}
