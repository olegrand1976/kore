import { describe, expect, it } from 'vitest'
import {
  formatLocalDate,
  initialActiveWeekNumber,
  weekNumberForDay
} from '../composables/useWeekCalendar'

describe('initialActiveWeekNumber', () => {
  it('returns the first month week when month is not current', () => {
    const now = new Date(2026, 8, 9) // 2026-09-09
    expect(initialActiveWeekNumber('2026-08', 1, now)).toBe(1)
  })

  it('returns the calendar week of today when month matches now', () => {
    const now = new Date(2026, 8, 9) // Wednesday 2026-09-09
    const today = formatLocalDate(now)
    const expected = weekNumberForDay('2026-09', today, 1)
    expect(initialActiveWeekNumber('2026-09', 1, now)).toBe(expected)
    expect(expected).toBeGreaterThan(1)
  })

  it('respects weekStartDay when resolving the current week', () => {
    const now = new Date(2026, 8, 9)
    const monStart = initialActiveWeekNumber('2026-09', 1, now)
    const sunStart = initialActiveWeekNumber('2026-09', 0, now)
    expect(monStart).toBe(weekNumberForDay('2026-09', '2026-09-09', 1))
    expect(sunStart).toBe(weekNumberForDay('2026-09', '2026-09-09', 0))
  })
})
