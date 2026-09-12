import { describe, expect, it } from 'vitest'
import {
  buildCreatePeriodMonthOptions,
  buildCreatePeriodYearOptions,
  clampCreatePeriodSelection,
  isStrictlyPastMonth,
  monthKeyFromParts,
  previousMonthKey,
  resolveOpenTimesheetUserId
} from '../utils/craCreatePeriod'

describe('craCreatePeriod', () => {
  const now = new Date(2026, 8, 15) // Sep 2026

  it('previousMonthKey returns prior calendar month', () => {
    expect(previousMonthKey(now)).toBe('2026-08')
    expect(previousMonthKey(new Date(2026, 0, 5))).toBe('2025-12')
  })

  it('isStrictlyPastMonth excludes current and future', () => {
    expect(isStrictlyPastMonth('2026-08', now)).toBe(true)
    expect(isStrictlyPastMonth('2026-09', now)).toBe(false)
    expect(isStrictlyPastMonth('2026-10', now)).toBe(false)
  })

  it('buildCreatePeriodMonthOptions limits current year to past months', () => {
    const months = buildCreatePeriodMonthOptions('2026', 'fr', now)
    expect(months.map((m) => m.value)).toEqual([
      '01', '02', '03', '04', '05', '06', '07', '08'
    ])
    expect(buildCreatePeriodMonthOptions('2025', 'fr', now)).toHaveLength(12)
    expect(buildCreatePeriodMonthOptions('2027', 'fr', now)).toEqual([])
  })

  it('clampCreatePeriodSelection falls back to previous month', () => {
    expect(clampCreatePeriodSelection('2026', '09', now)).toEqual({ year: '2026', month: '08' })
    expect(clampCreatePeriodSelection('2025', '03', now)).toEqual({ year: '2025', month: '03' })
    expect(monthKeyFromParts('2025', '03')).toBe('2025-03')
  })

  it('buildCreatePeriodYearOptions includes a past window', () => {
    expect(buildCreatePeriodYearOptions(now, 2).map((y) => y.value)).toEqual([
      '2026', '2025', '2024'
    ])
  })
})

describe('resolveOpenTimesheetUserId', () => {
  it('omits userId for non-validators and for self', () => {
    expect(resolveOpenTimesheetUserId(false, 'me', 'other')).toBe('')
    expect(resolveOpenTimesheetUserId(true, 'me', 'me')).toBe('')
    expect(resolveOpenTimesheetUserId(true, 'me', '')).toBe('')
  })

  it('returns filtered collaborator for validators', () => {
    expect(resolveOpenTimesheetUserId(true, 'me', 'gerald')).toBe('gerald')
  })
})
