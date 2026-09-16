import assert from 'node:assert/strict';
import test from 'node:test';

import en from '../../../src/locales/en.js';
import zh from '../../../src/locales/zh.js';
import { getSplitModeKey, SPLIT_MODES } from './split-mode.mjs';

const lookup = (messages, key) => key.split('.').reduce((value, part) => value?.[part], messages);

test('registered split modes have copy and unknown ones stay blank', () => {
  assert.equal(getSplitModeKey('mig'), 'card.splitMode.mig');
  assert.equal(lookup(zh, getSplitModeKey('hami-core')), '软切分（hami-core）');
  assert.equal(lookup(zh, getSplitModeKey('template')), '模板切分');
  for (const mode of SPLIT_MODES) {
    assert.equal(typeof lookup(en, getSplitModeKey(mode)), 'string', mode);
  }
  for (const mode of ['', undefined, 'something-new']) {
    assert.equal(getSplitModeKey(mode), '');
  }
});
