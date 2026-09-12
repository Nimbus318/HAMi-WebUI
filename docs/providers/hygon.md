# Hygon DCU and HCU compatibility

HAMi-WebUI keeps legacy DCU and current HCU integrations separate. Configure the
selector that matches your node labels:

```yaml
vendorNodeSelectors:
  DCU: dcu=on
  HCU: hygon.com/hcu=true
```

Direct backend configuration uses the same keys under `node_selectors`. An older
config file without `HCU` uses `hygon.com/hcu=true` for that provider. HCU devices
are discovered from `hami.io/node-hcu-register`, which the Hygon device plugin
writes in `strategy=hami`. Physical-only, pre-split and MIG plugin modes without
HAMi registration/allocation annotations are outside this integration.

The browser receives the actual provider (`DCU` or `HCU`) and registered model,
such as `HCU-K100_AI`. Both remain available in the existing inventory and model
filters; the frontend has no vendor allowlist requiring separate HCU UI code.

## Registration and allocation identities

| Contract | Legacy DCU | HCU |
| --- | --- | --- |
| Node registration | `hami.io/node-dcu-register` | `hami.io/node-hcu-register` |
| Pod allocation | `hami.io/dcu-devices-allocated` | `hami.io/hcu-devices-allocated` |
| Allocation identity | Node-local `DCU-<index>` mapped to `<node>-dcu-<index>` | Stable `HCU-<serial>` |
| Physical exporter identity | `dcu_temp{node,minor_number}` maps to `device_id` | Strip only the leading `HCU-` to obtain exporter `device_id`; preserve the full ID for allocation joins |

HCU registration includes device index, split count, schedulable memory in MiB,
core percentage, model, NUMA and health. It remains visible when telemetry is
absent. Pod decoding preserves init-container slots and whole-card allocations.

## HCU exporter contract

Use the default metric and label names from `HYGON-AI/hcu-exporter`. Its optional
`--metrics-define` and `--label-define` renaming is not inferred by WebUI. Preserve
the exporter `container` label when scraping (`honor_labels: true` when target
labels collide).

| Metric | Required identity | Value consumed by WebUI |
| --- | --- | --- |
| `hcu_memorycap_bytes`, `hcu_usedmemory_bytes` | `device_id` | Bytes, converted to MiB |
| `hcu_utilizationrate` | `device_id` | Physical-device utilization percent |
| `hcu_temp`, `hcu_temp_mem` | `device_id` | Degrees Celsius |
| `hcu_power_usage` | `device_id`, `minor_number` | Watts; minor number supplies the displayed device number |
| `vhcu_utilizationrate` | `device_id`, `node`, `hcu_pod_namespace`, `hcu_pod_name`, `container` | Virtual-device busy percent; active allocated vCore estimate is busy percent multiplied by allocated vCore / 100 |
| `vhcu_usedmemory_bytes` | Same workload identity | Bytes, converted to MiB |

For a HAMi whole-card allocation, the exporter attaches the same workload labels
to `hcu_utilizationrate` and `hcu_usedmemory_bytes`. WebUI uses those only when the
physical device, node, namespace, Pod and container all match and no virtual
sample exists. Unattributed whole-card telemetry is never used as container
usage. A missing workload sample stays unavailable; a reported zero is idle.

The exporter does not expose a Pod UID label. Workload identity therefore uses
its node/namespace/Pod/container labels, plus the physical device serial. Driver
version, fan and XID fields are not synthesized from unrelated metrics.

## Pinned source evidence

These source revisions define the compatibility fixtures; they are not a claim
of physical-hardware acceptance:

- [HAMi HCU adapter, `88b118e`](https://github.com/Project-HAMi/HAMi/blob/88b118e565a9effaa27f81669da291c5508fc5e4/pkg/device/hygon/device.go): registration, resource and allocation names.
- [Hygon device registration, `cc7d33f`](https://github.com/HYGON-AI/k8s-hcu-device-plugin/blob/cc7d33f4b00acca00b8657144aed72e4e2d5604f/internal/pkg/plugin/register.go): serial-prefixed ID, model, capacity and health; [wire encoding](https://github.com/HYGON-AI/k8s-hcu-device-plugin/blob/cc7d33f4b00acca00b8657144aed72e4e2d5604f/internal/pkg/util/util.go).
- [HCU exporter labels, `bfe780b`](https://github.com/HYGON-AI/hcu-exporter/blob/bfe780b89831a6245f19498c1068d17836d84653/cmd/hcu-exporter/metrics_loop.go): physical and virtual device/workload identities; [metric readers](https://github.com/HYGON-AI/hcu-exporter/blob/bfe780b89831a6245f19498c1068d17836d84653/cmd/hcu-exporter/main.go).
- [hcu-dcgm v3.0.0, `b223b72`](https://github.com/HYGON-AI/hcu-dcgm/blob/b223b72ee0c0079ad1b93a637fc96e00636f7a65/pkg/dcgm/api.go): serial-number identity, watt conversion and virtual-device busy percent.
