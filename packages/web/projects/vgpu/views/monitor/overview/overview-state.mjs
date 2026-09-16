import { REQUEST_STATUS } from '../../../../../src/hooks/request-state.mjs';

export const UNKNOWN_SHARES_STATUS = 'unknown-shares';

export const aggregateStatuses = (items = []) => {
  const statuses = items.map((item) => item?.status).filter(Boolean);
  if (statuses.includes(REQUEST_STATUS.READY)) return REQUEST_STATUS.READY;
  if (statuses.includes(REQUEST_STATUS.LOADING)) return REQUEST_STATUS.LOADING;
  if (statuses.includes(REQUEST_STATUS.ERROR)) return REQUEST_STATUS.ERROR;
  if (statuses.includes(REQUEST_STATUS.INVALID)) return REQUEST_STATUS.INVALID;
  if (statuses.includes('no-capacity')) return 'no-capacity';
  // An explained absence still beats a bare "no data".
  if (statuses.includes(UNKNOWN_SHARES_STATUS)) return UNKNOWN_SHARES_STATUS;
  return REQUEST_STATUS.MISSING;
};

// Compute allocation is suppressed while any allocation's share is unknown;
// say so instead of reporting missing data. Gauges carry an id, trend series a key.
export const applyUnknownShareStatus = (metrics = [], unknownCount = 0) => (
  metrics.map((metric) => (
    (metric?.id ?? metric?.key) === 'compute-allocation'
    && metric.status === REQUEST_STATUS.MISSING
    && unknownCount > 0
      ? { ...metric, status: UNKNOWN_SHARES_STATUS, unknownShares: unknownCount }
      : metric
  ))
);

export const stateTextKey = (status, { metric = true } = {}) => {
  if (status === REQUEST_STATUS.LOADING) return 'common.loading';
  if (status === REQUEST_STATUS.ERROR) {
    return metric ? 'dashboard.metricQueryFailed' : 'common.requestError';
  }
  if (status === REQUEST_STATUS.INVALID) {
    return metric ? 'dashboard.metricInvalid' : 'common.requestError';
  }
  if (status === 'no-capacity') return 'dashboard.metricNoCapacity';
  if (status === UNKNOWN_SHARES_STATUS) return 'dashboard.metricUnknownShares';
  return metric ? 'dashboard.metricNoData' : 'common.noData';
};

export const getPartialRangeStates = (items = []) => {
  if (aggregateStatuses(items) !== REQUEST_STATUS.READY) return [];
  return items.filter((item) => item?.status !== REQUEST_STATUS.READY);
};

export const selectRangeAxisData = (items = []) => {
  const hasData = (item) => Array.isArray(item?.data) && item.data.length > 0;
  return (
    items.find(
      (item) => item?.status === REQUEST_STATUS.READY && hasData(item),
    )?.data ||
    items.find(hasData)?.data ||
    []
  );
};
