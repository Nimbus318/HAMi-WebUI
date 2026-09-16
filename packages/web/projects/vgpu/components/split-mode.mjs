// How HAMi divides a device, as its device plugins register it.
export const SPLIT_MODES = ['hami-core', 'mig', 'template'];

export const getSplitModeKey = (mode) => (SPLIT_MODES.includes(mode) ? `card.splitMode.${mode}` : '');
