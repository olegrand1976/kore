export type WorkflowRecipientScope = 'user' | 'equipe' | 'service' | 'application' | 'all'

export type WorkflowSideEffectRecipients = {
  scope: WorkflowRecipientScope
  userIds?: string[]
  equipeId?: string
  serviceId?: string
  applicationId?: string
}

export type WorkflowSideEffect = {
  type: 'email'
  recipients: WorkflowSideEffectRecipients
  subject: string
  bodyTemplate: string
}

export type WorkflowState = {
  code: string
  label: string
  isInitial: boolean
  isFinal: boolean
  onEnterEffects?: WorkflowSideEffect[]
}

export type WorkflowTransition = {
  from: string
  to: string
  action: string
  guard?: string
  allowedRoles: string[]
  onFireEffects?: WorkflowSideEffect[]
}

export type WorkflowDefinition = {
  code: string
  entityType: string
  states: WorkflowState[]
  transitions: WorkflowTransition[]
}

export const WORKFLOW_RECIPIENT_SCOPES = ['user', 'equipe', 'service', 'application', 'all'] as const

export type WorkflowRecipientScopeOption = (typeof WORKFLOW_RECIPIENT_SCOPES)[number]

export const MAX_SIDE_EFFECTS_PER_HOOK = 10

export const WORKFLOW_PRESET_CODES: WorkflowPresetCode[] = ['leave.request', 'tma.incident']

export const WORKFLOW_ROLE_OPTIONS = ['Collaborateur', 'Administrateur', 'Utilisateur'] as const

export type WorkflowRoleOption = (typeof WORKFLOW_ROLE_OPTIONS)[number]

export type WorkflowPresetStateMeta = {
  code: string
  defaultLabel: string
  hintKey: string
  isInitial: boolean
  isFinal: boolean
}

export type WorkflowPresetTransitionMeta = {
  from: string
  to: string
  action: string
  labelKey: string
  hintKey: string
  screenKey: string
  defaultRoles: string[]
}

export type WorkflowPresetMeta = {
  code: WorkflowPresetCode
  entityType: string
  labelKey: string
  descKey: string
  howtoKey: string
  summaryKey: string
  states: WorkflowPresetStateMeta[]
  transitions: WorkflowPresetTransitionMeta[]
}

export const WORKFLOW_PRESET_META: Record<WorkflowPresetCode, WorkflowPresetMeta> = {
  'leave.request': {
    code: 'leave.request',
    entityType: 'leave_request',
    labelKey: 'workflows.preset_leave',
    descKey: 'workflows.preset_leave_desc',
    howtoKey: 'workflows.howto.leave_example',
    summaryKey: 'workflows.assistant.summary_leave',
    states: [
      {
        code: 'en_attente',
        defaultLabel: 'En attente',
        hintKey: 'workflows.assistant.states.en_attente',
        isInitial: true,
        isFinal: false
      },
      {
        code: 'valide',
        defaultLabel: 'Validé',
        hintKey: 'workflows.assistant.states.valide',
        isInitial: false,
        isFinal: true
      },
      {
        code: 'refuse',
        defaultLabel: 'Refusé',
        hintKey: 'workflows.assistant.states.refuse',
        isInitial: false,
        isFinal: true
      }
    ],
    transitions: [
      {
        from: 'en_attente',
        to: 'valide',
        action: 'approve',
        labelKey: 'workflows.assistant.actions.approve',
        hintKey: 'workflows.assistant.hints.leave_approve',
        screenKey: 'workflows.assistant.screens.conges_validation',
        defaultRoles: []
      },
      {
        from: 'en_attente',
        to: 'refuse',
        action: 'reject',
        labelKey: 'workflows.assistant.actions.reject',
        hintKey: 'workflows.assistant.hints.leave_reject',
        screenKey: 'workflows.assistant.screens.conges_validation',
        defaultRoles: []
      }
    ]
  },
  'tma.incident': {
    code: 'tma.incident',
    entityType: 'tma_demand',
    labelKey: 'workflows.preset_tma',
    descKey: 'workflows.preset_tma_desc',
    howtoKey: 'workflows.howto.tma_example',
    summaryKey: 'workflows.assistant.summary_tma',
    states: [
      {
        code: 'en_attente_creation',
        defaultLabel: 'En attente création',
        hintKey: 'workflows.assistant.states.en_attente_creation',
        isInitial: false,
        isFinal: false
      },
      {
        code: 'ouverte',
        defaultLabel: 'Ouverte',
        hintKey: 'workflows.assistant.states.ouverte',
        isInitial: true,
        isFinal: false
      },
      {
        code: 'affectee',
        defaultLabel: 'Affectée',
        hintKey: 'workflows.assistant.states.affectee',
        isInitial: false,
        isFinal: false
      },
      {
        code: 'resolue',
        defaultLabel: 'Résolue',
        hintKey: 'workflows.assistant.states.resolue',
        isInitial: false,
        isFinal: true
      },
      {
        code: 'rework',
        defaultLabel: 'Rework',
        hintKey: 'workflows.assistant.states.rework',
        isInitial: false,
        isFinal: false
      }
    ],
    transitions: [
      {
        from: 'en_attente_creation',
        to: 'ouverte',
        action: 'validate_creation',
        labelKey: 'workflows.assistant.actions.validate_creation',
        hintKey: 'workflows.assistant.hints.tma_validate_creation',
        screenKey: 'workflows.assistant.screens.tma_detail',
        defaultRoles: []
      },
      {
        from: 'ouverte',
        to: 'affectee',
        action: 'assign',
        labelKey: 'workflows.assistant.actions.assign',
        hintKey: 'workflows.assistant.hints.tma_assign',
        screenKey: 'workflows.assistant.screens.tma_detail',
        defaultRoles: []
      },
      {
        from: 'affectee',
        to: 'resolue',
        action: 'resolve',
        labelKey: 'workflows.assistant.actions.resolve',
        hintKey: 'workflows.assistant.hints.tma_resolve',
        screenKey: 'workflows.assistant.screens.tma_detail',
        defaultRoles: []
      },
      {
        from: 'resolue',
        to: 'rework',
        action: 'reopen',
        labelKey: 'workflows.assistant.actions.reopen',
        hintKey: 'workflows.assistant.hints.tma_reopen',
        screenKey: 'workflows.assistant.screens.tma_detail',
        defaultRoles: []
      },
      {
        from: 'rework',
        to: 'affectee',
        action: 'assign',
        labelKey: 'workflows.assistant.actions.assign',
        hintKey: 'workflows.assistant.hints.tma_assign_rework',
        screenKey: 'workflows.assistant.screens.tma_detail',
        defaultRoles: []
      }
    ]
  }
}

/** @deprecated Utiliser WORKFLOW_PRESET_META — conservé pour compatibilité pages existantes */
export const WORKFLOW_PRESETS: Record<
  WorkflowPresetCode,
  { code: WorkflowPresetCode; entityType: string; labelKey: string; descKey: string; howtoKey: string }
> = {
  'leave.request': WORKFLOW_PRESET_META['leave.request'],
  'tma.incident': WORKFLOW_PRESET_META['tma.incident']
}

type RawWorkflowSideEffect = {
  type?: string
  Type?: string
  recipients?: RawWorkflowSideEffectRecipients
  Recipients?: RawWorkflowSideEffectRecipients
  subject?: string
  Subject?: string
  bodyTemplate?: string
  BodyTemplate?: string
}

type RawWorkflowSideEffectRecipients = {
  scope?: string
  Scope?: string
  userIds?: string[]
  UserIDs?: string[]
  equipeId?: string
  EquipeID?: string
  serviceId?: string
  ServiceID?: string
  applicationId?: string
  ApplicationID?: string
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
  onEnterEffects?: RawWorkflowSideEffect[]
  OnEnterEffects?: RawWorkflowSideEffect[]
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
  onFireEffects?: RawWorkflowSideEffect[]
  OnFireEffects?: RawWorkflowSideEffect[]
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
  | 'side_effect_invalid'
  | 'side_effect_subject_required'
  | 'side_effect_recipients_required'
  | 'too_many_side_effects'

function normalizeRecipientScope(raw: string | undefined): WorkflowRecipientScope {
  switch (raw) {
    case 'user':
    case 'equipe':
    case 'service':
    case 'application':
    case 'all':
      return raw
    default:
      return 'user'
  }
}

function normalizeSideEffectRecipients(raw: RawWorkflowSideEffectRecipients | undefined): WorkflowSideEffectRecipients {
  const scope = normalizeRecipientScope(raw?.scope ?? raw?.Scope)
  return {
    scope,
    userIds: raw?.userIds ?? raw?.UserIDs ?? [],
    equipeId: raw?.equipeId ?? raw?.EquipeID,
    serviceId: raw?.serviceId ?? raw?.ServiceID,
    applicationId: raw?.applicationId ?? raw?.ApplicationID
  }
}

function normalizeSideEffect(raw: RawWorkflowSideEffect): WorkflowSideEffect {
  const type = raw.type ?? raw.Type ?? 'email'
  if (type !== 'email') {
    return {
      type: 'email',
      recipients: { scope: 'user', userIds: [] },
      subject: '',
      bodyTemplate: ''
    }
  }
  return {
    type: 'email',
    recipients: normalizeSideEffectRecipients(raw.recipients ?? raw.Recipients),
    subject: raw.subject ?? raw.Subject ?? '',
    bodyTemplate: raw.bodyTemplate ?? raw.BodyTemplate ?? ''
  }
}

function normalizeSideEffects(raw: RawWorkflowSideEffect[] | undefined): WorkflowSideEffect[] {
  return (raw ?? []).map(normalizeSideEffect)
}

function validateSideEffects(effects: WorkflowSideEffect[] | undefined): WorkflowValidationCode[] {
  const errors: WorkflowValidationCode[] = []
  const list = effects ?? []
  if (list.length > MAX_SIDE_EFFECTS_PER_HOOK) {
    errors.push('too_many_side_effects')
    return errors
  }
  for (const effect of list) {
    if (effect.type !== 'email') {
      errors.push('side_effect_invalid')
      continue
    }
    if (!effect.subject.trim() && !effect.bodyTemplate.trim()) {
      errors.push('side_effect_subject_required')
    }
    switch (effect.recipients.scope) {
      case 'all':
        break
      case 'user':
        if (!(effect.recipients.userIds?.length ?? 0)) errors.push('side_effect_recipients_required')
        break
      case 'equipe':
        if (!effect.recipients.equipeId?.trim()) errors.push('side_effect_recipients_required')
        break
      case 'service':
        if (!effect.recipients.serviceId?.trim()) errors.push('side_effect_recipients_required')
        break
      case 'application':
        if (!effect.recipients.applicationId?.trim()) errors.push('side_effect_recipients_required')
        break
      default: {
        const _exhaustive: never = effect.recipients.scope
        errors.push('side_effect_invalid')
        void _exhaustive
      }
    }
  }
  return errors
}

function transitionKey(tr: Pick<WorkflowTransition, 'from' | 'action' | 'to'>): string {
  return `${tr.from}|${tr.action}|${tr.to}`
}

function rolesEqual(a: string[], b: string[]): boolean {
  const sa = [...a].sort().join(',')
  const sb = [...b].sort().join(',')
  return sa === sb
}

export function buildPresetDefinition(code: WorkflowPresetCode): WorkflowDefinition {
  const meta = WORKFLOW_PRESET_META[code]
  return {
    code: meta.code,
    entityType: meta.entityType,
    states: meta.states.map((s) => ({
      code: s.code,
      label: s.defaultLabel,
      isInitial: s.isInitial,
      isFinal: s.isFinal
    })),
    transitions: meta.transitions.map((tr) => ({
      from: tr.from,
      to: tr.to,
      action: tr.action,
      allowedRoles: [...tr.defaultRoles]
    }))
  }
}

/** Réapplique la structure preset en conservant libellés et rôles chargés depuis l'API. */
export function mergePresetWithLoaded(code: WorkflowPresetCode, loaded: WorkflowDefinition): WorkflowDefinition {
  const preset = buildPresetDefinition(code)
  const labelByCode = new Map(loaded.states.map((s) => [s.code, s.label]))
  const effectsByState = new Map(loaded.states.map((s) => [s.code, s.onEnterEffects ?? []]))
  const rolesByKey = new Map(loaded.transitions.map((tr) => [transitionKey(tr), tr.allowedRoles]))
  const effectsByTransition = new Map(
    loaded.transitions.map((tr) => [transitionKey(tr), tr.onFireEffects ?? []])
  )

  return {
    ...preset,
    states: preset.states.map((s) => ({
      ...s,
      label: labelByCode.get(s.code)?.trim() || s.label,
      onEnterEffects: effectsByState.get(s.code) ?? []
    })),
    transitions: preset.transitions.map((tr) => ({
      ...tr,
      allowedRoles: rolesByKey.get(transitionKey(tr)) ?? tr.allowedRoles,
      onFireEffects: effectsByTransition.get(transitionKey(tr)) ?? []
    }))
  }
}

export function differsFromPreset(editor: WorkflowDefinition): boolean {
  if (!isPresetCode(editor.code)) return false
  const preset = buildPresetDefinition(editor.code)

  for (const ps of preset.states) {
    const es = editor.states.find((s) => s.code === ps.code)
    if (!es || es.label.trim() !== ps.label.trim()) return true
  }

  for (const pt of preset.transitions) {
    const et = editor.transitions.find(
      (tr) => tr.from === pt.from && tr.action === pt.action && tr.to === pt.to
    )
    if (!et || !rolesEqual(pt.allowedRoles, et.allowedRoles)) return true
  }

  return false
}

export function getPresetMeta(code: WorkflowPresetCode): WorkflowPresetMeta {
  return WORKFLOW_PRESET_META[code]
}

export function getStateMeta(code: WorkflowPresetCode, stateCode: string): WorkflowPresetStateMeta | undefined {
  return WORKFLOW_PRESET_META[code].states.find((s) => s.code === stateCode)
}

export function getTransitionMeta(
  code: WorkflowPresetCode,
  tr: Pick<WorkflowTransition, 'from' | 'action' | 'to'>
): WorkflowPresetTransitionMeta | undefined {
  return WORKFLOW_PRESET_META[code].transitions.find(
    (t) => t.from === tr.from && t.action === tr.action && t.to === tr.to
  )
}

export function normalizeDefinition(raw: RawWorkflowDefinition, fallbackCode: string): WorkflowDefinition {
  const preset = WORKFLOW_PRESET_META[fallbackCode as WorkflowPresetCode]
  const normalized: WorkflowDefinition = {
    code: raw.code ?? raw.Code ?? fallbackCode,
    entityType: raw.entityType ?? raw.EntityType ?? preset?.entityType ?? '',
    states: (raw.states ?? raw.States ?? []).map((s) => ({
      code: s.code ?? s.Code ?? '',
      label: s.label ?? s.Label ?? '',
      isInitial: s.isInitial ?? s.IsInitial ?? false,
      isFinal: s.isFinal ?? s.IsFinal ?? false,
      onEnterEffects: normalizeSideEffects(s.onEnterEffects ?? s.OnEnterEffects)
    })),
    transitions: (raw.transitions ?? raw.Transitions ?? []).map((tr) => ({
      from: tr.from ?? tr.From ?? '',
      to: tr.to ?? tr.To ?? '',
      action: tr.action ?? tr.Action ?? '',
      guard: tr.guard ?? tr.Guard ?? '',
      allowedRoles: tr.allowedRoles ?? tr.AllowedRoles ?? [],
      onFireEffects: normalizeSideEffects(tr.onFireEffects ?? tr.OnFireEffects)
    }))
  }

  if (isPresetCode(fallbackCode)) {
    return mergePresetWithLoaded(fallbackCode, normalized)
  }
  return normalized
}

export function buildPayload(definition: WorkflowDefinition): WorkflowDefinition {
  return {
    code: definition.code,
    entityType: definition.entityType,
    states: definition.states.map((s) => ({
      code: s.code.trim(),
      label: s.label.trim(),
      isInitial: s.isInitial,
      isFinal: s.isFinal,
      onEnterEffects: (s.onEnterEffects ?? []).map(sanitizeSideEffect)
    })),
    transitions: definition.transitions.map((tr) => ({
      from: tr.from.trim(),
      to: tr.to.trim(),
      action: tr.action.trim(),
      guard: tr.guard?.trim() ?? '',
      allowedRoles: [...tr.allowedRoles],
      onFireEffects: (tr.onFireEffects ?? []).map(sanitizeSideEffect)
    }))
  }
}

function sanitizeSideEffect(effect: WorkflowSideEffect): WorkflowSideEffect {
  return {
    type: 'email',
    recipients: {
      scope: effect.recipients.scope,
      userIds: effect.recipients.userIds?.filter(Boolean),
      equipeId: effect.recipients.equipeId?.trim() || undefined,
      serviceId: effect.recipients.serviceId?.trim() || undefined,
      applicationId: effect.recipients.applicationId?.trim() || undefined
    },
    subject: effect.subject.trim(),
    bodyTemplate: effect.bodyTemplate.trim()
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
    errors.push(...validateSideEffects(tr.onFireEffects))
  }

  for (const state of definition.states) {
    errors.push(...validateSideEffects(state.onEnterEffects))
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
  return { code: '', label: '', isInitial: false, isFinal: false, onEnterEffects: [] }
}

export function createEmptyTransition(): WorkflowTransition {
  return { from: '', to: '', action: '', guard: '', allowedRoles: [], onFireEffects: [] }
}

export function createEmptySideEffect(): WorkflowSideEffect {
  return {
    type: 'email',
    recipients: { scope: 'all' },
    subject: '',
    bodyTemplate: ''
  }
}

export function useWorkflowDefinition() {
  return {
    WORKFLOW_PRESET_CODES,
    WORKFLOW_PRESETS,
    WORKFLOW_PRESET_META,
    WORKFLOW_ROLE_OPTIONS,
    WORKFLOW_RECIPIENT_SCOPES,
    MAX_SIDE_EFFECTS_PER_HOOK,
    normalizeDefinition,
    buildPresetDefinition,
    mergePresetWithLoaded,
    differsFromPreset,
    getPresetMeta,
    getStateMeta,
    getTransitionMeta,
    buildPayload,
    validateDefinition,
    isPresetCode,
    stateReferencedByTransition,
    createEmptyState,
    createEmptyTransition,
    createEmptySideEffect
  }
}
