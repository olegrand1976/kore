import { describe, expect, it } from 'vitest'
import {
  coerceBudgetDefautId,
  defaultBudgetsForApplication,
  filterByApplicationId,
  isDefaultBudgetType,
  pickAppActive,
  pickAppClient,
  pickAppEquipeIds,
  pickAppId,
  pickAppLabel,
  pickAppMode,
  pickAppServiceIds,
  pickAppSiteIds,
  summarizeAppShares
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

  it('reads share id lists and summarizes them', () => {
    expect(pickAppServiceIds({ serviceIds: ['s1', 's2'] })).toEqual(['s1', 's2'])
    expect(pickAppServiceIds({ ServiceIDs: ['s3'] })).toEqual(['s3'])
    expect(pickAppServiceIds({ serviceId: 'legacy' })).toEqual(['legacy'])
    expect(pickAppSiteIds({ siteIds: ['site-1'] })).toEqual(['site-1'])
    expect(pickAppEquipeIds({ EquipeIDs: ['e1'] })).toEqual(['e1'])
    expect(summarizeAppShares({ siteIds: ['a'], serviceIds: ['b', 'c'], equipeIds: [] })).toEqual({
      sites: 1,
      services: 2,
      equipes: 0
    })
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

  it('clears stale budgetDefautId outside allowed default budgets', () => {
    expect(coerceBudgetDefautId('', ['d1'])).toBe('')
    expect(coerceBudgetDefautId('d1', ['d1', 'd2'])).toBe('d1')
    expect(coerceBudgetDefautId('stale', ['d1'])).toBe('')
  })
})
