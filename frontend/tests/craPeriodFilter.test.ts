import { describe, expect, it } from 'vitest'
import {
  buildMonthFilterOptions,
  buildYearFilterOptions,
  matchPeriodMonthYear
} from '../utils/craPeriodFilter'

describe('matchPeriodMonthYear', () => {
  it('matches when both filters are empty', () => {
    expect(matchPeriodMonthYear('2026-09', '', '')).toBe(true)
  })

  it('filters by year only', () => {
    expect(matchPeriodMonthYear('2026-09', '', '2026')).toBe(true)
    expect(matchPeriodMonthYear('2025-09', '', '2026')).toBe(false)
  })

  it('filters by month only', () => {
    expect(matchPeriodMonthYear('2026-09', '09', '')).toBe(true)
    expect(matchPeriodMonthYear('2026-08', '09', '')).toBe(false)
  })

  it('filters by month and year', () => {
    expect(matchPeriodMonthYear('2026-09', '09', '2026')).toBe(true)
    expect(matchPeriodMonthYear('2026-08', '09', '2026')).toBe(false)
  })
})

describe('buildYearFilterOptions', () => {
  it('includes current year and unique years from keys', () => {
    const opts = buildYearFilterOptions(['2026-09', '2025-01', '2026-01'], new Date(2026, 8, 9))
    expect(opts.map((o) => o.value)).toEqual(['2026', '2025'])
  })
})

describe('buildMonthFilterOptions', () => {
  it('returns 12 months', () => {
    expect(buildMonthFilterOptions('fr')).toHaveLength(12)
    expect(buildMonthFilterOptions('fr')[0]?.value).toBe('01')
    expect(buildMonthFilterOptions('fr')[11]?.value).toBe('12')
  })
})
