<template>
  <div class="device-split">
    <div class="device-split-bar" role="img" :aria-label="barLabel">
      <div
        v-for="segment in split.segments"
        :key="segment.key"
        class="device-split-segment"
        :class="{ 'is-free': segment.kind === 'free', 'is-current': segment.highlighted }"
        :style="{ flexGrow: segment.size, flexBasis: 0 }"
      >
        <span v-if="segment.kind === 'free'" class="device-split-free">{{ $t('card.split.free') }}</span>
        <template v-else>
          <span class="device-split-name">{{ segment.label || memoryText(segment.memoryMiB) }}</span>
          <span class="device-split-owner">{{ workloadName(segment.workload) }}</span>
        </template>
      </div>
    </div>
    <ul class="device-split-list">
      <li v-for="segment in allocations" :key="segment.key" :class="{ 'is-current': segment.highlighted }">
        <span class="device-split-dot" :class="{ 'is-current': segment.highlighted }" aria-hidden="true" />
        <span class="device-split-item-name">{{ segment.label || $t('card.split.noTemplate') }}</span>
        <span class="device-split-item-size">{{ sizeText(segment) }}</span>
        <RouterLink
          v-if="workloadLocation(segment.workload)"
          class="device-split-item-workload"
          :to="workloadLocation(segment.workload)"
        >{{ workloadName(segment.workload) }}</RouterLink>
        <span v-else class="device-split-item-workload">{{ workloadName(segment.workload) }}</span>
        <span v-if="segment.highlighted" class="device-split-current">{{ $t('card.split.thisWorkload') }}</span>
      </li>
    </ul>
    <p v-if="!allocations.length" class="device-split-empty">{{ $t('card.split.empty') }}</p>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { RouterLink } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { roundToDecimal } from '@/utils';
import { buildDeviceSplit } from '../device-split.mjs';
import { buildWorkloadDetailLocation } from '~/vgpu/views/task/admin/workload-identity.mjs';

const props = defineProps({
  device: { type: Object, required: true },
  containers: { type: Array, default: () => [] },
  highlight: { type: Object, default: undefined },
});
const { t } = useI18n();

const split = computed(() => buildDeviceSplit({ device: props.device, containers: props.containers, highlight: props.highlight }));
const allocations = computed(() => split.value.segments.filter((segment) => segment.kind === 'allocation'));
const memoryText = (mib) => `${roundToDecimal(Number(mib || 0) / 1024, 2)} GiB`;
const sizeText = (segment) => (split.value.scale === 'slices'
  ? t('card.split.slices', { count: segment.size, memory: memoryText(segment.memoryMiB) })
  : memoryText(segment.memoryMiB));
const workloadName = ({ appName, container } = {}) => [appName, container].filter(Boolean).join(' / ') || '--';
const workloadLocation = (workload = {}) => buildWorkloadDetailLocation({ podUid: workload.podUid, name: workload.container });
const barLabel = computed(() => t('card.split.title'));
</script>

<style lang="scss" scoped>
.device-split-bar {
  display: flex;
  gap: 4px;
  width: 100%;
  margin: 12px 0;
}

.device-split-segment {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 2px;
  min-width: 0;
  padding: 12px 10px;
  overflow: hidden;
  border-radius: 6px;
  background: #e8effb;
  color: #2f66e0;

  &.is-free {
    background: #f5f7fa;
    color: #939ea9;
  }

  &.is-current {
    background: #2f66e0;
    color: #fff;
  }
}

.device-split-name,
.device-split-free {
  font-size: 13px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.device-split-owner {
  font-size: 12px;
  opacity: 0.8;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.device-split-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin: 0;
  padding: 0;
  list-style: none;

  li {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    color: #5f6b7a;
  }
}

.device-split-dot {
  width: 8px;
  height: 8px;
  border-radius: 2px;
  background: #e8effb;

  &.is-current {
    background: #2f66e0;
  }
}

.device-split-item-name {
  min-width: 96px;
  color: #324558;
  font-weight: 500;
}

.device-split-item-size {
  min-width: 120px;
}

.device-split-item-workload {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #2f66e0;
}

.device-split-current {
  padding: 0 6px;
  border-radius: 10px;
  background: #e8effb;
  color: #2f66e0;
  line-height: 18px;
}

.device-split-empty {
  margin: 0;
  color: #939ea9;
  font-size: 12px;
}
</style>
