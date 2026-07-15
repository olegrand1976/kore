export type WorkflowState = {
  code: string
  label: string
  isInitial: boolean
  isFinal: boolean
}

export type WorkflowTransition = {
  from: string
  to: string
  action: string
  guard?: string
  allowedRoles: string[]
}

export type WorkflowDefinition = {
  code: string
  entityType: string
  states: WorkflowState[]
  transitions: WorkflowTransition[]
}

export type WorkflowPresetCode = 'leave.request' | 'tma.incident'

export const WORKFLOW_PRESET_CODES: WorkflowPresetCode[] = ['leave.request', 'tma.incident']

export const WORKFLOW_ROLE_OPTIONS = ['Collaborateur', 'Administrateur', 'Utilisateur'] as const

export type WorkflowRoleOption = (typeof WORKFLOW_ROLE_OPTIONS)[number]

export const WORKFLOW_PRESETS: Record<
  WorkflowPresetCode,
  { code: WorkflowPresetCode; entityType: string; labelKey: string; descKey: string; howtoKey: string }
> = {
  'leave.request': {
    code: 'leave.request',
    entityType: 'leave_request',
    labelKey: 'workflows.preset_leave',
    descKey: 'workflows.preset_leave_desc',
    howtoKey: 'workflows.howto.leave_example'
  },
  'tma.incident': {
    code: 'tma.incident',
    entityType: 'tma_demand',
    labelKey: 'workflows.preset_tma',
    descKey: 'workflows.preset_tma_desc',
    howtoKey: 'workflows.howto.tma_example'
  }
}

type RawWorkflowState = {
  code?: string
  Code?: string
  label?: string
  Label?: string
  isInitial?: boolean
  IsInitial?: boolean
  isFinal?: boolean
  IsFinal?: boolean
}

type RawWorkflowTransition = {
  from?: string
  From?: string
  to?: string
  To?: string
  action?: string
  Action?: string
  guard?: string
  Guard?: string
  allowedRoles?: string[]
  AllowedRoles?: string[]
}

export type RawWorkflowDefinition = {
  code?: string
  Code?: string
  entityType?: string
  EntityType?: string
  states?: RawWorkflowState[]
  States?: RawWorkflowState[]
  transitions?: RawWorkflowTransition[]
  Transitions?: RawWorkflowTransition[]
}

export type WorkflowValidationCode =
  | 'code_required'
  | 'entity_type_required'
  | 'states_required'
  | 'one_initial'
  | 'one_final'
  | 'duplicate_state'
  | 'orphan_transition'
  | 'transition_action_required'

export function buildPresetDefinition(code: WorkflowPresetCode): WorkflowDefinition {
  switch (code) {
    case 'leave.request':
      return {
        code,
        entityType: 'leave_request',
        states: [
          { code: 'en_attente', label: 'En attente', isInitial: true, isFinal: false },
          { code: 'valide', label: 'Validé', isInitial: false, isFinal: true },
          { code: 'refuse', label: 'Refusé', isInitial: false, isFinal: true }
        ],
        transitions: [
          { from: 'en_attente', to: 'valide', action: 'approve', allowedRoles: [] },
          { from: 'en_attente', to: 'refuse', action: 'reject', allowedRoles: [] }
        ]
      }
    case 'tma.incident':
      return {
        code,
        entityType: 'tma_demand',
        states: [
          { code: 'en_attente_creation', label: 'En attente création', isInitial: false, isFinal: false },
          { code: 'ouverte', label: 'Ouverte', isInitial: true, isFinal: false },
          { code: 'affectee', label: 'Affectée', isInitial: false, isFinal: false },
          { code: 'resolue', label: 'Résolue', isInitial: false, isFinal: true },
          { code: 'rework', label: 'Rework', isInitial: false, isFinal: false }
        ],
        transitions: [
          { from: 'en_attente_creation', to: 'ouverte', action: 'validate_creation', allowedRoles: [] },
          { from: 'ouverte', to: 'affectee', action: 'assign', allowedRoles: [] },
          { from: 'affectee', to: 'resolue', action: 'resolve', allowedRoles: [] },
          { from: 'resolue', to: 'rework', action: 'reopen', allowedRoles: [] },
          { from: 'rework', to: 'affectee', action: 'assign', allowedRoles: [] }
        ]
      }
    default: {
      const _exhaustive: never = code
      return _exhaustive
    }
  }
}

export function normalizeDefinition(raw: RawWorkflowDefinition, fallbackCode: string): WorkflowDefinition {
  const preset = WORKFLOW_PRESETS[fallbackCode as WorkflowPresetCode]
  return {
    code: raw.code ?? raw.Code ?? fallbackCode,
    entityType: raw.entityType ?? raw.EntityType ?? preset?.entityType ?? '',
    states: (raw.states ?? raw.States ?? []).map((s) => ({
      code: s.code ?? s.Code ?? '',
      label: s.label ?? s.Label ?? '',
      isInitial: s.isInitial ?? s.IsInitial ?? false,
      isFinal: s.isFinal ?? s.IsFinal ?? false
    })),
    transitions: (raw.transitions ?? raw.Transitions ?? []).map((tr) => ({
      from: tr.from ?? tr.From ?? '',
      to: tr.to ?? tr.To ?? '',
      action: tr.action ?? tr.Action ?? '',
      guard: tr.guard ?? tr.Guard ?? '',
      allowedRoles: tr.allowedRoles ?? tr.AllowedRoles ?? []
    }))
  }
}

export function buildPayload(definition: WorkflowDefinition): WorkflowDefinition {
  return {
    code: definition.code,
    entityType: definition.entityType,
    states: definition.states.map((s) => ({
      code: s.code.trim(),
      label: s.label.trim(),
      isInitial: s.isInitial,
      isFinal: s.isFinal
    })),
    transitions: definition.transitions.map((tr) => ({
      from: tr.from.trim(),
      to: tr.to.trim(),
      action: tr.action.trim(),
      guard: tr.guard?.trim() ?? '',
      allowedRoles: [...tr.allowedRoles]
    }))
  }
}

export function validateDefinition(definition: WorkflowDefinition): WorkflowValidationCode[] {
  const errors: WorkflowValidationCode[] = []

  if (!definition.code.trim()) errors.push('code_required')
  if (!definition.entityType.trim()) errors.push('entity_type_required')
  if (definition.states.length === 0) errors.push('states_required')

  const initialCount = definition.states.filter((s) => s.isInitial).length
  if (initialCount !== 1) errors.push('one_initial')

  const finalCount = definition.states.filter((s) => s.isFinal).length
  if (finalCount === 0) errors.push('one_final')

  const stateCodes = definition.states.map((s) => s.code.trim()).filter(Boolean)
  if (new Set(stateCodes).size !== stateCodes.length) errors.push('duplicate_state')

  const stateSet = new Set(stateCodes)
  for (const tr of definition.transitions) {
    if (!tr.action.trim()) {
      errors.push('transition_action_required')
      continue
    }
    if (!stateSet.has(tr.from.trim()) || !stateSet.has(tr.to.trim())) {
      errors.push('orphan_transition')
    }
  }

  return [...new Set(errors)]
}

export function isPresetCode(code: string): code is WorkflowPresetCode {
  return WORKFLOW_PRESET_CODES.includes(code as WorkflowPresetCode)
}

export function stateReferencedByTransition(
  code: string,
  transitions: WorkflowTransition[]
): boolean {
  return transitions.some((tr) => tr.from === code || tr.to === code)
}

export function createEmptyState(): WorkflowState {
  return { code: '', label: '', isInitial: false, isFinal: false }
}

export function createEmptyTransition(): WorkflowTransition {
  return { from: '', to: '', action: '', guard: '', allowedRoles: [] }
}

export function useWorkflowDefinition() {
  return {
    WORKFLOW_PRESET_CODES,
    WORKFLOW_PRESETS,
    WORKFLOW_ROLE_OPTIONS,
    normalizeDefinition,
    buildPresetDefinition,
    buildPayload,
    validateDefinition,
    isPresetCode,
    stateReferencedByTransition,
    createEmptyState,
    createEmptyTransition
  }
}
