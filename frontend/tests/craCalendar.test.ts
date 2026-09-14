import { describe, expect, it } from 'vitest'
import {
  collectIncompleteWorkingDays,
  countWorkingDays,
  isWeekend
} from '../utils/craCalendar'

describe('isWeekend', () => {
  it('detects Saturday and Sunday', () => {
    expect(isWeekend('2026-08-22')).toBe(true) // Saturday
    expect(isWeekend('2026-08-23')).toBe(true) // Sunday
    expect(isWeekend('2026-08-21')).toBe(false) // Friday
  })
})

describe('countWorkingDays', () => {
  it('excludes weekends', () => {
    expect(
      countWorkingDays([
        '2026-08-17',
        '2026-08-18',
        '2026-08-19',
        '2026-08-20',
        '2026-08-21',
        '2026-08-22',
        '2026-08-23'
      ])
    ).toBe(5)
  })
})

describe('collectIncompleteWorkingDays', () => {
  const toMinutes = (hours: string) => {
    const n = Number(hours)
    return Number.isFinite(n) ? Math.round(n * 60) : 0
  }

  it('ignores Saturday and Sunday without hours', () => {
    const days = ['2026-08-17', '2026-08-18', '2026-08-19', '2026-08-20', '2026-08-21', '2026-08-22', '2026-08-23']
    const rows = new Map(
      days.map((day) => [
        day,
        day <= '2026-08-21' ? [{ hours: '8', sourceType: 'manual' }] : []
      ])
    )
    expect(collectIncompleteWorkingDays(days, rows, toMinutes)).toEqual([])
  })

  it('skips holiday source days', () => {
    const days = ['2026-07-14']
    const rows = new Map([['2026-07-14', [{ hours: '', sourceType: 'holiday' }]]])
    expect(collectIncompleteWorkingDays(days, rows, toMinutes)).toEqual([])
  })

  it('flags empty working days', () => {
    const days = ['2026-08-17', '2026-08-18']
    const rows = new Map([
      ['2026-08-17', [{ hours: '8', sourceType: 'manual' }]],
      ['2026-08-18', []]
    ])
    expect(collectIncompleteWorkingDays(days, rows, toMinutes)).toEqual(['2026-08-18'])
  })
})
