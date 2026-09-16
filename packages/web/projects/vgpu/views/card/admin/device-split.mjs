const positive = (value) => {
  const number = Number(value);
  return Number.isFinite(number) && number > 0 ? number : 0;
};

const migSpace = (device) => (device.migProfiles || []).reduce((space, profile) => (
  (profile.placements || []).reduce((widest, placement) => Math.max(widest, positive(placement.start) + positive(placement.size)), space)
), 0);

const identity = (container) => ({
  podUid: container.podUid,
  container: container.name,
  appName: container.appName,
  namespace: container.namespace,
});

const sameWorkload = (left, right) => Boolean(left && right && left.podUid === right.podUid && left.container === right.container);

// The split of one device: what HAMi placed on it, and what is still free.
// MIG devices are laid out in the GPU's placement space, everything else by memory.
export const buildDeviceSplit = ({ device = {}, containers = [], highlight } = {}) => {
  const entries = [];
  for (const container of containers) {
    for (const allocated of container.devices || []) {
      if (allocated.id && allocated.id === device.uuid) entries.push({ container, allocated });
    }
  }
  const placed = entries.filter(({ allocated }) => positive(allocated.migSize) > 0);
  const scale = placed.length === entries.length && placed.length > 0 ? 'slices' : 'memory';
  return scale === 'slices'
    ? sliceSplit(device, placed, highlight)
    : memorySplit(device, entries, highlight);
};

const sliceSplit = (device, entries, highlight) => {
  const sorted = [...entries].sort((left, right) => positive(left.allocated.migStart) - positive(right.allocated.migStart));
  const used = sorted.reduce((end, { allocated }) => Math.max(end, positive(allocated.migStart) + positive(allocated.migSize)), 0);
  const total = Math.max(migSpace(device), used);
  const segments = [];
  let cursor = 0;
  for (const { container, allocated } of sorted) {
    const start = positive(allocated.migStart);
    const size = positive(allocated.migSize);
    if (start > cursor) segments.push({ kind: 'free', key: `free-${cursor}`, size: start - cursor });
    segments.push({
      kind: 'allocation',
      key: `${container.podUid}/${container.name}/${start}`,
      size,
      label: allocated.template,
      memoryMiB: positive(allocated.allocatedMem),
      cores: allocated.allocatedCores,
      workload: identity(container),
      highlighted: sameWorkload(identity(container), highlight),
    });
    cursor = Math.max(cursor, start + size);
  }
  if (total > cursor) segments.push({ kind: 'free', key: `free-${cursor}`, size: total - cursor });
  return { scale: 'slices', total, segments };
};

const memorySplit = (device, entries, highlight) => {
  const total = positive(device.memoryTotal);
  const segments = entries.map(({ container, allocated }) => ({
    kind: 'allocation',
    key: `${container.podUid}/${container.name}/${allocated.id}`,
    size: positive(allocated.allocatedMem),
    label: allocated.template,
    memoryMiB: positive(allocated.allocatedMem),
    cores: allocated.allocatedCores,
    workload: identity(container),
    highlighted: sameWorkload(identity(container), highlight),
  }));
  const used = segments.reduce((sum, segment) => sum + segment.size, 0);
  if (total > used) segments.push({ kind: 'free', key: `free-${used}`, size: total - used });
  return { scale: 'memory', total: Math.max(total, used), segments };
};
