const positive = (value) => {
  const number = Number(value);
  return Number.isFinite(number) && number > 0 ? number : undefined;
};

export const DEVICE_CONFIG_STATES = ['disabled', 'loading', 'missing', 'forbidden', 'invalid', 'error'];

// int64 fields arrive as strings; absent AI Core counts stay undefined.
export const findAscendModel = (config, type) => {
  if (config?.state !== 'loaded' || !type) return undefined;
  const model = (Array.isArray(config.ascendModels) ? config.ascendModels : [])
    .find((item) => item?.commonWord === type);
  if (!model) return undefined;
  return {
    commonWord: model.commonWord,
    chipName: model.chipName || '',
    aiCore: positive(model.aiCore),
    aiCpu: positive(model.aiCpu),
    memoryAllocatableMiB: positive(model.memoryAllocatable),
    templates: (Array.isArray(model.templates) ? model.templates : []).map((template) => ({
      name: template.name,
      memoryMiB: positive(template.memory),
      aiCore: positive(template.aiCore),
      computeShare: positive(template.computeShare),
    })),
  };
};

export const getDeviceConfigStateKey = (config) => {
  const state = config?.state;
  if (state === 'loaded') return '';
  return `card.deviceConfig.state.${DEVICE_CONFIG_STATES.includes(state) ? state : 'error'}`;
};
