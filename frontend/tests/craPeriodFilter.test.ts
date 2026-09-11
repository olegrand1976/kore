import { describe, expect, it } from 'vitest'
import {
  buildMonthFilterOptions,
  buildYearFilterOptions,
  matchPeriodMonthYear,
  migrateLegacyMonthFilter
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
  it('includes a window around current year plus years from keys', () => {
    const opts = buildYearFilterOptions(['2026-09', '2025-01', '2026-01'], new Date(2026, 8, 9))
    expect(opts.map((o) => o.value)).toEqual(['2027', '2026', '2025', '2024'])
  })
})

describe('buildMonthFilterOptions', () => {
  it('returns 12 months', () => {
    expect(buildMonthFilterOptions('fr')).toHaveLength(12)
    expect(buildMonthFilterOptions('fr')[0]?.value).toBe('01')
    expect(buildMonthFilterOptions('fr')[11]?.value).toBe('12')
  })
})

describe('migrateLegacyMonthFilter', () => {
  it('splits YYYY-MM into periodYear and periodMonth', () => {
    const filters: Record<string, string> = { month: '2026-09' }
    expect(migrateLegacyMonthFilter(filters)).toBe(true)
    expect(filters).toEqual({ periodYear: '2026', periodMonth: '09' })
  })

  it('removes invalid legacy month key', () => {
    const filters: Record<string, string> = { month: 'invalid', status: '' }
    expect(migrateLegacyMonthFilter(filters)).toBe(false)
    expect(filters).toEqual({ status: '' })
  })
})
