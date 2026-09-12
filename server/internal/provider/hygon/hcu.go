package hygon

import (
	"fmt"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"vgpu/internal/provider/util"
)

// HCU consumes the HAMi-mode registration of k8s-hcu-device-plugin. Legacy
// DCU registration uses a different device identity and remains in Hygon.
type HCU struct {
	log           *log.Helper
	labelSelector string
}

func NewHCU(logger *log.Helper, labelSelector string) *HCU {
	if labelSelector == "" {
		// Existing direct config files may not yet contain the new HCU key.
		labelSelector = "hygon.com/hcu=true"
	}
	return &HCU{log: logger, labelSelector: labelSelector}
}

func (h *HCU) GetNodeDevicePluginLabels() (labels.Selector, error) {
	return labels.Parse(h.labelSelector)
}

func (h *HCU) GetProvider() string {
	return HygonHCUDevice
}

func (h *HCU) FetchDevices(node *corev1.Node) ([]*util.DeviceInfo, error) {
	encoded, ok := node.Annotations[HCURegisterAnnos]
	if !ok {
		return []*util.DeviceInfo{}, nil
	}
	devices, err := util.DecodeNodeDevices(encoded, h.log)
	if err != nil {
		return nil, err
	}
	for _, device := range devices {
		// The plugin prefixes the physical serial with HCU- in both Node and
		// Pod annotations. hcu-exporter exposes that same serial as device_id;
		// unlike legacy DCU, this is not a node-local minor number.
		serial, ok := strings.CutPrefix(device.ID, "HCU-")
		if !ok || serial == "" {
			return nil, fmt.Errorf("invalid HCU registration ID %q on node %s", device.ID, node.Name)
		}
		device.AliasId = device.ID
		device.ID = serial
	}
	return devices, nil
}
