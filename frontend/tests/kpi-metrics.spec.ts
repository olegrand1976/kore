import { beforeEach, describe, expect, it, vi } from 'vitest'

const MONTH = '2026-08'

vi.mock('../composables/useCraStatus', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../composables/useCraStatus')>()
  return {
    ...actual,
    currentMonthKey: () => MONTH
  }
})

import { isCraMonthIncomplete } from '../composables/useKpiMetrics'

describe('isCraMonthIncomplete', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('marks missing current month as incomplete', () => {
    expect(isCraMonthIncomplete([])).toBe(true)
    expect(
      isCraMonthIncomplete([
        { month: '2000-01', status: 'Définitif', weeksSubmitted: 4, weeksTotal: 4, totalMinutes: 480 }
      ])
    ).toBe(true)
  })

  it('marks Définitif as complete even with zero minutes', () => {
    expect(
      isCraMonthIncomplete([
        { month: MONTH, status: 'Définitif', weeksSubmitted: 0, weeksTotal: 4, totalMinutes: 0 }
      ])
    ).toBe(false)
  })

  it('marks incomplete when submitted weeks are below total', () => {
    expect(
      isCraMonthIncomplete([
        { month: MONTH, status: 'ValidéSemaine', weeksSubmitted: 1, weeksTotal: 4, totalMinutes: 960 }
      ])
    ).toBe(true)
  })

  it('marks incomplete when minutes are zero with weeks filled', () => {
    expect(
      isCraMonthIncomplete([
        { month: MONTH, status: 'Brouillon', weeksSubmitted: 4, weeksTotal: 4, totalMinutes: 0 }
      ])
    ).toBe(true)
  })

  it('falls back to minutes when weeksTotal is zero', () => {
    expect(
      isCraMonthIncomplete([
        { month: MONTH, status: 'Brouillon', weeksSubmitted: 0, weeksTotal: 0, totalMinutes: 0 }
      ])
    ).toBe(true)
    expect(
      isCraMonthIncomplete([
        { month: MONTH, status: 'Brouillon', weeksSubmitted: 0, weeksTotal: 0, totalMinutes: 480 }
      ])
    ).toBe(false)
  })

  it('marks complete when weeks are filled and minutes are positive', () => {
    expect(
      isCraMonthIncomplete([
        { month: MONTH, status: 'ValidéSemaine', weeksSubmitted: 4, weeksTotal: 4, totalMinutes: 480 }
      ])
    ).toBe(false)
  })

  it('accepts PascalCase API fields', () => {
    expect(
      isCraMonthIncomplete([
        {
          Month: MONTH,
          Status: 'ValidéSemaine',
          WeeksSubmitted: 2,
          WeeksTotal: 4,
          TotalMinutes: 100
        }
      ])
    ).toBe(true)
  })
})
