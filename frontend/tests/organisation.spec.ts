import { describe, expect, it } from 'vitest'
import {
  buildEquipeOptions,
  equipesForApplication,
  formatEquipeOptionLabels,
  orgId,
  orgLabel,
  planEquipeMembershipUpdates,
  unwrapOrgData
} from '~/composables/useOrganisation'

describe('org payload normalisation', () => {
  // Les handlers Go renvoient tantôt les tags json camelCase, tantôt les noms de
  // champs Go exportés : les deux formes doivent être lues indifféremment.
  it('reads ids in both casings', () => {
    expect(orgId({ id: 'a' })).toBe('a')
    expect(orgId({ ID: 'b' })).toBe('b')
    expect(orgId(undefined)).toBe('')
  })

  it('reads labels in both casings', () => {
    expect(orgLabel({ libelle: 'Équipe Dev' })).toBe('Équipe Dev')
    expect(orgLabel({ Libelle: 'Équipe TMA' })).toBe('Équipe TMA')
    expect(orgLabel(undefined)).toBe('')
  })
})

describe('buildEquipeOptions', () => {
  it('maps each team to a value/label pair', () => {
    const options = buildEquipeOptions([
      { id: 'e1', applicationId: 'a1', libelle: 'Équipe Dev' }
    ])
    expect(options).toEqual([{ value: 'e1', label: 'Équipe Dev' }])
  })

  it('keeps homonymous teams as separate options', () => {
    const options = buildEquipeOptions([
      { id: 'e1', applicationId: 'a1', libelle: 'Équipe Dev' },
      { id: 'e2', applicationId: 'a2', libelle: 'Équipe Dev' }
    ])
    expect(options).toEqual([
      { value: 'e1', label: 'Équipe Dev' },
      { value: 'e2', label: 'Équipe Dev' }
    ])
  })

  it('handles Go-cased payloads', () => {
    const options = buildEquipeOptions([
      { ID: 'e1', ApplicationID: 'a1', Libelle: 'Équipe Dev' }
    ])
    expect(options).toEqual([{ value: 'e1', label: 'Équipe Dev' }])
  })

  it('falls back to the team id when the label is empty', () => {
    expect(buildEquipeOptions([{ id: 'e1', applicationId: 'a1' }])).toEqual([
      { value: 'e1', label: 'e1' }
    ])
  })

  it('returns an empty list when there is no team', () => {
    expect(buildEquipeOptions([])).toEqual([])
  })
})

describe('equipesForApplication', () => {
  const equipes = [
    { id: 'e1', applicationId: 'a1', libelle: 'MG Consulting' },
    { id: 'e2', applicationId: 'a2', libelle: 'SOFT-CONNECT' },
    { id: 'e3', applicationId: 'a3', libelle: 'Support' }
  ]

  it('returns owned and shared teams sorted by label', () => {
    expect(equipesForApplication(equipes, 'a1', ['e3'])).toEqual([
      { value: 'e1', label: 'MG Consulting' },
      { value: 'e3', label: 'Support' }
    ])
  })

  it('returns an empty list when the application id is missing', () => {
    expect(equipesForApplication(equipes, '', ['e1'])).toEqual([])
  })

  it('handles Go-cased payloads', () => {
    expect(
      equipesForApplication(
        [{ ID: 'e1', ApplicationID: 'a1', Libelle: 'MG Consulting' }],
        'a1',
        ['e1']
      )
    ).toEqual([{ value: 'e1', label: 'MG Consulting' }])
  })

  it('does not duplicate an owned team also listed as shared', () => {
    expect(equipesForApplication(equipes, 'a1', ['e1', 'e3'])).toEqual([
      { value: 'e1', label: 'MG Consulting' },
      { value: 'e3', label: 'Support' }
    ])
  })
})

describe('formatEquipeOptionLabels', () => {
  it('joins team labels for display', () => {
    expect(
      formatEquipeOptionLabels([
        { value: 'e1', label: 'MG Consulting' },
        { value: 'e2', label: 'SOFT-CONNECT' }
      ])
    ).toBe('MG Consulting, SOFT-CONNECT')
  })

  it('returns an empty string when there are no teams', () => {
    expect(formatEquipeOptionLabels([])).toBe('')
  })
})

describe('planEquipeMembershipUpdates', () => {
  const users = [
    { userId: 'u1', equipeIds: ['e1'] },
    { userId: 'u2', equipeIds: [] },
    { userId: 'u3', equipeIds: ['e1', 'e2'] }
  ]

  it('adds and removes members to match the desired set', () => {
    expect(planEquipeMembershipUpdates('e1', ['u2', 'u3'], users)).toEqual([
      { userId: 'u1', equipeIds: [] },
      { userId: 'u2', equipeIds: ['e1'] }
    ])
  })

  it('forces the responsable into membership', () => {
    expect(
      planEquipeMembershipUpdates('e1', ['u1'], users, { ensureUserId: 'u2' })
    ).toEqual([
      { userId: 'u2', equipeIds: ['e1'] },
      { userId: 'u3', equipeIds: ['e2'] }
    ])
  })

  it('returns nothing when already in sync', () => {
    expect(planEquipeMembershipUpdates('e1', ['u1', 'u3'], users)).toEqual([])
  })

  it('ignores empty equipe id', () => {
    expect(planEquipeMembershipUpdates('', ['u1'], users)).toEqual([])
  })
})

describe('unwrapOrgData', () => {
  it('unwraps data envelopes and passes through bare payloads', () => {
    expect(unwrapOrgData<{ id: string }>({ data: { id: 'x' } })).toEqual({ id: 'x' })
    expect(unwrapOrgData<{ id: string }>({ id: 'y' })).toEqual({ id: 'y' })
  })
})
