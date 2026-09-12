# Experimental Moore Threads sGPU compatibility

This branch layers a bounded Moore Threads adapter onto the HCU compatibility
work. It is a lab integration branch, not an upstream release or a claim that
vendor drivers executed in the lab. No frontend allowlist change is required:
the existing views display the returned provider and model with a generic GPU
icon when no vendor icon exists.

The implementation was informed by the design reviewed in
[WebUI PR 287](https://github.com/Project-HAMi/HAMi-WebUI/pull/287), head
`e165b89b0abbd70f96e70447b0a48b40257a87ec`, authored in
`wjluo/HAMi-WebUI`. That PR was open and unmerged when checked on 2026-09-12.
This implementation does not import its whole-card occupancy ledger, global
Summary service, or rule that interprets zero reserved cores as a whole card.
Provider registration, allocation decoding and telemetry mapping are bounded
to the supported sGPU path described below.

## Source contracts

- [HAMi Mthreads adapter at 88b118e565a9effaa27f81669da291c5508fc5e4](https://github.com/Project-HAMi/HAMi/blob/88b118e565a9effaa27f81669da291c5508fc5e4/pkg/device/mthreads/device.go):
  node capacity `mthreads.com/sgpu-core` uses 16 units per card;
  `mthreads.com/sgpu-memory` uses 512 MiB per unit; card IDs are
  `<node>-mthreads-<index>`; retained allocated annotations contain MiB and
  the original 0..16 reserved core weight.
- [Official MT DCGM Exporter guide](https://docs.mthreads.com/cloud-native/cloud-native-doc-online/user_guide/exporter_guide/):
  GPU utilization %, framebuffer memory MB, temperatures Celsius and power W.
- [Official installation example](https://docs.mthreads.com/cloud-native/cloud-native-doc-online/install_guide/operator_install/):
  raw labels include uppercase `Hostname`, numeric `gpu`, `device=mtgpu<N>`,
  `UUID`, `modelName` and version fields.
- [Current release notes](https://docs.mthreads.com/cloud-native/cloud-native-doc-online/releasenote/):
  KUAE 2.2.0, MT DCGM Exporter `1.1.7-3.3.6-3.4.2`. The installation example
  still shows older driver versions; the adapter targets the documented profile,
  not a source-verified or hardware-captured instance of that binary.

## Supported behavior

Set `vendorNodeSelectors.Mthreads` to `mthreads.com/gpu-node=true`, or to the
deployment's actual selector. The lab uses `lab.hami.io/vendor=mthreads` only
for node selection. The adapter does not read the simulator's inventory JSON.

The provider reconstructs homogeneous, contiguous sGPU inventory from native
Kubernetes capacities. It preserves HAMi's scheduler IDs and 100 allocation
slots per card. WebUI percentage fields normalize 16 reserved core units to
100; zero stays zero. The integer API has one percentage-point resolution, so
fractions such as 1/16 are truncated to 6. Memory stays in MiB. No whole-card
Pod acquires an invented allocation when a retained annotation is absent.

Physical queries match `Hostname`, `gpu` and `device=mtgpu<N>`. The raw exporter
UUID need not equal HAMi's scheduler ID. The model and driver are optional
display metadata; missing Prometheus samples preserve scheduler inventory.
The `Health` inventory field follows HAMi's hard-coded healthy state and is not
a hardware-health observation. Framebuffer MB values follow the binary-unit
convention of the synthetic lab; this conversion is not independently confirmed
against a live MT SDK.

Only physical usage is emitted. Workload allocation series come from retained
HAMi annotations, with an exact card-ID match. Workload usage returns the
existing unsupported error and makes no physical-card fallback query. Missing
physical samples remain absent, real zeros remain present, and transport errors
degrade the existing collection status.

Unsupported paths include mixed whole-card/sGPU index layouts, heterogeneous
capacities within one node, allocation without HAMi retained annotations,
independent sGPU workload utilization/memory, and inferred XID or health enums.
The global Summary behavior is unchanged.

## Verification

Run `GOTOOLCHAIN=go1.26.7 make -C server verify`. The provider tests cover native
capacity/identity, metadata enrichment, missing collection and excluded whole
cards. Decoder tests cover zero cores, the 16-unit scale, MiB, container slots
and retained versus pending allocation. Exporter tests cover device labels and
units, zero versus missing, collection failure, exact allocation joins and the
absence of invented workload measurements.
