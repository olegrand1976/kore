import { describe, expect, it } from 'vitest'
import {
  hoursInputToMinutes,
  minutesToHoursInput,
  resolveWeekCapacityMinutes
} from '../utils/craWeekCapacity'

describe('resolveWeekCapacityMinutes', () => {
  it('uses planned week minutes when set', () => {
    expect(
      resolveWeekCapacityMinutes({
        weekDayCount: 5,
        dayCapacityMinutes: 480,
        plannedWeekMinutes: 2100
      })
    ).toBe(2100)
  })

  it('falls back to days × day capacity', () => {
    expect(
      resolveWeekCapacityMinutes({
        weekDayCount: 4,
        dayCapacityMinutes: 480,
        plannedWeekMinutes: null
      })
    ).toBe(1920)
  })
})

describe('hours ↔ minutes helpers', () => {
  it('converts hours input to minutes', () => {
    expect(hoursInputToMinutes('35')).toBe(2100)
    expect(hoursInputToMinutes('37.5')).toBe(2250)
    expect(hoursInputToMinutes('')).toBeNull()
  })

  it('formats minutes for the hours input', () => {
    expect(minutesToHoursInput(2100)).toBe('35')
    expect(minutesToHoursInput(2250)).toBe('37.5')
    expect(minutesToHoursInput(null)).toBe('')
  })
})
