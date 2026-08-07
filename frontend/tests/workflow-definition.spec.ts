import { describe, expect, it } from 'vitest'
import {
  WORKFLOW_PRESET_CODES,
  WORKFLOW_RECIPIENT_SCOPES,
  buildPresetDefinition,
  validateDefinition
} from '../composables/useWorkflowDefinition'

describe('useWorkflowDefinition', () => {
  it('exposes requester recipient scope', () => {
    expect(WORKFLOW_RECIPIENT_SCOPES).toContain('requester')
  })

  it('includes invoicing.cra_proforma preset with valid structure', () => {
    expect(WORKFLOW_PRESET_CODES).toContain('invoicing.cra_proforma')
    const def = buildPresetDefinition('invoicing.cra_proforma')
    expect(def.code).toBe('invoicing.cra_proforma')
    expect(def.entityType).toBe('invoice')
    expect(validateDefinition(def)).toEqual([])
    expect(def.states.find((s) => s.code === 'proforma_refusee')?.isFinal).toBe(false)
    expect(def.transitions.filter((tr) => tr.action === 'emit_proforma').map((tr) => tr.from).sort()).toEqual([
      'preparee',
      'proforma',
      'proforma_refusee'
    ])
    expect(def.transitions.some((tr) => tr.action === 'validate_client')).toBe(true)
    expect(def.transitions.some((tr) => tr.action === 'reject_client')).toBe(true)
  })

  it('accepts requester scope without picker ids', () => {
    const def = buildPresetDefinition('leave.request')
    def.states[0].onEnterEffects = [
      {
        type: 'email',
        recipients: { scope: 'requester' },
        subject: 'Demande reçue',
        bodyTemplate: ''
      }
    ]
    expect(validateDefinition(def)).toEqual([])
  })
})
