# MLU sGPU inventory units

The lab compatibility branch uses the current HAMi scheduler's native sMLU
inventory contract. In [HAMi 88b118e565a9effaa27f81669da291c5508fc5e4](https://github.com/Project-HAMi/HAMi/blob/88b118e565a9effaa27f81669da291c5508fc5e4/pkg/device/cambricon/device.go),
each card has 100 allocation slots, and the `vmemory` resource has a 256 MiB unit.
WebUI previously used 10 slots and 1024 MiB; the latter inflated schedulable
memory fourfold. Two cards exposing 200 vcore and 384 vmemory units must show
100 slots and 49152 MiB per card.

Physical exporter memory remains bytes; it is a different contract and is not
rescaled at the producer to compensate for a consumer error. The regression test
uses a synthetic native `mlu_health` Prometheus response for identity and the
native node capacities for allocation inventory.

This adjustment targets the existing `mlu370.smlu.*` resource-name profile.
It does not introduce arbitrary model-resource discovery or change the global
summary service.

The scheduler emits `CAMBRICON_DSMLU_PROFILE=slot_core_memoryUnits`. WebUI now
decodes that reservation directly. The distinct Device Plugin result
`CAMBRICON_DSMLU_PROFILE_INSTANCE` has four fields
`profileID_handle_slot_instanceID` in [Cambricon Device Plugin v2.3.0](https://github.com/Cambricon/cambricon-k8s-device-plugin/blob/5731364a23f4938fbe42c150b51de3f16eaf605d/device-plugin/pkg/mlu/server.go#L450).
It is not appended to the profile or used as a guessed memory field. Before
runtime completion that result is absent; the old concatenation indexed beyond
the four-element intermediate string and could crash WebUI. The new decoder
validates the three scheduler fields, preserves zero reserved cores, and works
both before and after the native four-field runtime result arrives.
