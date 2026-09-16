import { REQUEST_STATUS } from '../../../src/hooks/request-state.mjs';

// How many allocations a compute allocation rate leaves out: a number once the
// count is read (a count that matches nothing returns no series, a confirmed
// zero), undefined while it loads, and null when it cannot be read.
export const readUncounted = (status, value) => {
  if (status === REQUEST_STATUS.READY) {
    const count = Number(value);
    return Number.isFinite(count) && count > 0 ? count : 0;
  }
  if (status === REQUEST_STATUS.MISSING) return 0;
  if (status === REQUEST_STATUS.LOADING) return undefined;
  return null;
};

// One sample that left allocations out makes the whole series a lower bound.
export const readUncountedRange = ({ status, data } = {}) => readUncounted(
  status,
  Math.max(0, ...(data || []).map((point) => Number(point?.value)).filter(Number.isFinite)),
);

// The rate never overstates, so it reads as a lower bound unless zero is confirmed.
export const isLowerBound = (uncounted) => uncounted === null || uncounted > 0;

export const lowerBoundMessage = (t, uncounted) => (uncounted === null
  ? t('dashboard.metricLowerBoundUnknown')
  : t('dashboard.metricLowerBound', { count: uncounted }));
