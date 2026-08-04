import { describe, expect, it } from 'vitest'
import { buildEquipeOptions, orgId, orgLabel } from '~/composables/useOrganisation'

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
  it('qualifies each team with its application', () => {
    const options = buildEquipeOptions(
      [{ id: 'e1', applicationId: 'a1', libelle: 'Équipe Dev' }],
      [{ id: 'a1', libelle: 'Portail Client' }]
    )
    expect(options).toEqual([{ value: 'e1', label: 'Équipe Dev — Portail Client' }])
  })

  it('disambiguates homonymous teams across applications', () => {
    const options = buildEquipeOptions(
      [
        { id: 'e1', applicationId: 'a1', libelle: 'Équipe Dev' },
        { id: 'e2', applicationId: 'a2', libelle: 'Équipe Dev' }
      ],
      [
        { id: 'a1', libelle: 'Portail Client' },
        { id: 'a2', libelle: 'Refonte ERP' }
      ]
    )
    expect(options.map((o) => o.label)).toEqual([
      'Équipe Dev — Portail Client',
      'Équipe Dev — Refonte ERP'
    ])
  })

  it('falls back to the bare team label when the application is unknown', () => {
    const options = buildEquipeOptions(
      [{ id: 'e1', applicationId: 'missing', libelle: 'Équipe orpheline' }],
      [{ id: 'a1', libelle: 'Portail Client' }]
    )
    expect(options).toEqual([{ value: 'e1', label: 'Équipe orpheline' }])
  })

  it('handles Go-cased payloads', () => {
    const options = buildEquipeOptions(
      [{ ID: 'e1', ApplicationID: 'a1', Libelle: 'Équipe Dev' }],
      [{ ID: 'a1', Libelle: 'Portail Client' }]
    )
    expect(options).toEqual([{ value: 'e1', label: 'Équipe Dev — Portail Client' }])
  })

  it('returns an empty list when there is no team', () => {
    expect(buildEquipeOptions([], [{ id: 'a1', libelle: 'Portail Client' }])).toEqual([])
  })
})
