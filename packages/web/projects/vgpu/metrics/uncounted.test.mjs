import assert from 'node:assert/strict';
import test from 'node:test';

import { isLowerBound, lowerBoundMessage, readUncounted, readUncountedRange } from './uncounted.mjs';

test('only a read count confirms how many allocations a rate leaves out', () => {
  assert.equal(readUncounted('ready', '2'), 2);
  assert.equal(readUncounted('ready', '0'), 0);
  // count() over nothing returns no series.
  assert.equal(readUncounted('missing'), 0);
  assert.equal(readUncounted('loading'), undefined);
  assert.equal(readUncounted('error'), null);
  assert.equal(readUncounted('invalid'), null);
});

test('a range is a lower bound when any sample left allocations out', () => {
  const points = (...values) => values.map((value, i) => ({ timestamp: i, value }));
  assert.equal(readUncountedRange({ status: 'ready', data: points(0, 1, null, 0) }), 1);
  assert.equal(readUncountedRange({ status: 'ready', data: points(0, 0) }), 0);
  assert.equal(readUncountedRange({ status: 'missing', data: [] }), 0);
  assert.equal(readUncountedRange({ status: 'error', data: [] }), null);
});

test('an unread count keeps the rate a lower bound without naming a number', () => {
  assert.equal(isLowerBound(0), false);
  assert.equal(isLowerBound(undefined), false);
  assert.equal(isLowerBound(3), true);
  assert.equal(isLowerBound(null), true);
  const t = (key, params) => `${key}${params ? `:${params.count}` : ''}`;
  assert.equal(lowerBoundMessage(t, 3), 'dashboard.metricLowerBound:3');
  assert.equal(lowerBoundMessage(t, null), 'dashboard.metricLowerBoundUnknown');
});
