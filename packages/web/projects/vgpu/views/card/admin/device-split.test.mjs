import assert from 'node:assert/strict';
import test from 'node:test';

import { buildDeviceSplit } from './device-split.mjs';

const a100 = {
  uuid: 'GPU-0',
  mode: 'mig',
  memoryTotal: 40960,
  migProfiles: [
    { name: '1g.5gb', placements: [{ start: 0, size: 1 }, { start: 1, size: 1 }, { start: 7, size: 1 }] },
    { name: '3g.20gb', placements: [{ start: 0, size: 4 }, { start: 4, size: 4 }] },
  ],
};
const container = (podUid, name, device) => ({ podUid, name, appName: `${podUid}-pod`, namespace: 'ml', devices: [device] });

test('MIG devices lay out in the placement space with the gaps between them', () => {
  const split = buildDeviceSplit({
    device: a100,
    containers: [
      container('pod-b', 'main', { id: 'GPU-0', template: '1g.5gb', migStart: 4, migSize: 1, allocatedMem: 5120, allocatedCores: 14 }),
      container('pod-a', 'main', { id: 'GPU-0', template: '3g.20gb', migStart: 0, migSize: 4, allocatedMem: 20480, allocatedCores: 43 }),
    ],
    highlight: { podUid: 'pod-b', container: 'main' },
  });
  assert.equal(split.scale, 'slices');
  assert.equal(split.total, 8);
  assert.deepEqual(split.segments.map(({ kind, size, label }) => [kind, size, label ?? null]), [
    ['allocation', 4, '3g.20gb'],
    ['allocation', 1, '1g.5gb'],
    ['free', 3, null],
  ]);
  assert.deepEqual(split.segments.map(({ highlighted }) => highlighted ?? false), [false, true, false]);
  assert.equal(split.segments[0].workload.appName, 'pod-a-pod');
});

test('other devices lay out by memory, with the remainder free', () => {
  const card = { uuid: 'NPU-0', mode: 'template', memoryTotal: 65536 };
  const split = buildDeviceSplit({
    device: card,
    containers: [
      container('pod-a', 'main', { id: 'NPU-0', template: 'vir05_1c_16g', allocatedMem: 16384, allocatedCores: 25 }),
      container('pod-b', 'worker', { id: 'other-card', allocatedMem: 32768 }),
    ],
  });
  assert.equal(split.scale, 'memory');
  assert.equal(split.total, 65536);
  assert.deepEqual(split.segments.map(({ kind, size }) => [kind, size]), [['allocation', 16384], ['free', 49152]]);
});

test('an empty device is all free and a MIG device without placements falls back to memory', () => {
  assert.deepEqual(buildDeviceSplit({ device: a100, containers: [] }).segments, [{ kind: 'free', key: 'free-0', size: 40960 }]);
  const noPlacement = buildDeviceSplit({
    device: a100,
    containers: [container('pod-a', 'main', { id: 'GPU-0', template: '1g.5gb', allocatedMem: 5120 })],
  });
  assert.equal(noPlacement.scale, 'memory');
  assert.deepEqual(buildDeviceSplit().segments, []);
});
