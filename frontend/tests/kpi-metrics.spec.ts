import { describe, expect, it } from 'vitest'
import { currentMonthKey } from '../composables/useCraStatus'
import { isCraMonthIncomplete } from '../composables/useKpiMetrics'

describe('isCraMonthIncomplete', () => {
  const month = currentMonthKey()

  it('marks missing current month as incomplete', () => {
    expect(isCraMonthIncomplete([])).toBe(true)
    expect(
      isCraMonthIncomplete([{ month: '2000-01', status: 'Définitif', weeksSubmitted: 4, weeksTotal: 4, totalMinutes: 480 }])
    ).toBe(true)
  })

  it('marks Définitif as complete even with zero minutes', () => {
    expect(
      isCraMonthIncomplete([
        { month, status: 'Définitif', weeksSubmitted: 0, weeksTotal: 4, totalMinutes: 0 }
      ])
    ).toBe(false)
  })

  it('marks incomplete when submitted weeks are below total', () => {
    expect(
      isCraMonthIncomplete([
        { month, status: 'ValidéSemaine', weeksSubmitted: 1, weeksTotal: 4, totalMinutes: 960 }
      ])
    ).toBe(true)
  })

  it('marks incomplete when minutes are zero with weeks filled', () => {
    expect(
      isCraMonthIncomplete([
        { month, status: 'Brouillon', weeksSubmitted: 4, weeksTotal: 4, totalMinutes: 0 }
      ])
    ).toBe(true)
  })

  it('marks complete when weeks are filled and minutes are positive', () => {
    expect(
      isCraMonthIncomplete([
        { month, status: 'ValidéSemaine', weeksSubmitted: 4, weeksTotal: 4, totalMinutes: 480 }
      ])
    ).toBe(false)
  })

  it('accepts PascalCase API fields', () => {
    expect(
      isCraMonthIncomplete([
        {
          Month: month,
          Status: 'ValidéSemaine',
          WeeksSubmitted: 2,
          WeeksTotal: 4,
          TotalMinutes: 100
        }
      ])
    ).toBe(true)
  })
})
