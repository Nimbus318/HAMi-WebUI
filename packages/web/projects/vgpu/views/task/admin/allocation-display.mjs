export const ALLOCATION_SHAPES = ['whole', 'template', 'soft', 'unknown'];
export const CORES_UNKNOWN_REASONS = [
  'catalog_unavailable',
  'model_not_configured',
  'template_not_configured',
  'compute_not_configured',
  'mode_ambiguous',
  'node_mode_unknown',
];

// undefined for allocations the server does not interpret.
export const getAllocationShapeCopy = ({ allocationShape, template } = {}) => {
  if (!ALLOCATION_SHAPES.includes(allocationShape)) return undefined;
  if (allocationShape === 'template' && template) {
    return { key: 'task.allocation.shape.templateNamed', params: { template } };
  }
  return { key: `task.allocation.shape.${allocationShape}`, params: {} };
};

export const getCoresUnknownReasonKey = ({ allocatedCoresKnown, allocatedCoresReason } = {}) => (
  allocatedCoresKnown === false && CORES_UNKNOWN_REASONS.includes(allocatedCoresReason)
    ? `task.allocation.reason.${allocatedCoresReason}`
    : ''
);
