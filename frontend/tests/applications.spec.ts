import { describe, expect, it } from 'vitest'
import {
  defaultBudgetsForApplication,
  filterByApplicationId,
  isDefaultBudgetType,
  pickAppActive,
  pickAppClient,
  pickAppId,
  pickAppLabel,
  pickAppMode
} from '~/composables/useApplications'

describe('useApplications pickers', () => {
  it('reads ids and labels in both casings', () => {
    expect(pickAppId({ id: 'a1' })).toBe('a1')
    expect(pickAppId({ ID: 'a2' })).toBe('a2')
    expect(pickAppLabel({ libelle: 'Portail' })).toBe('Portail')
    expect(pickAppLabel({ Libelle: 'ERP' })).toBe('ERP')
    expect(pickAppClient({ proprietaire: 'ACME' })).toBe('ACME')
    expect(pickAppMode({ modeFacturation: 'forfait' })).toBe('forfait')
    expect(pickAppMode({})).toBe('temps_passe')
    expect(pickAppActive({ active: false })).toBe(false)
    expect(pickAppActive({})).toBe(true)
  })

  it('filters items by applicationId', () => {
    const items = [
      { id: '1', applicationId: 'app-a' },
      { id: '2', ApplicationID: 'app-b' },
      { id: '3', applicationId: 'app-a' }
    ]
    expect(filterByApplicationId(items, 'app-a').map((i) => i.id)).toEqual(['1', '3'])
  })

  it('keeps only default-type budgets for an application (RG-BUD-01)', () => {
    expect(isDefaultBudgetType('defaut')).toBe(true)
    expect(isDefaultBudgetType('Default')).toBe(true)
    expect(isDefaultBudgetType('specifique')).toBe(false)
    const items = [
      { id: 'd1', applicationId: 'app-a', type: 'defaut' },
      { id: 's1', applicationId: 'app-a', type: 'specifique' },
      { id: 'd2', ApplicationID: 'app-a', Type: 'default' },
      { id: 'd3', applicationId: 'app-b', type: 'defaut' }
    ]
    expect(defaultBudgetsForApplication(items, 'app-a').map((i) => i.id)).toEqual(['d1', 'd2'])
  })
})
