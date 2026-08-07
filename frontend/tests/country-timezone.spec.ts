import { describe, expect, it } from 'vitest'
import {
  formatDualClock,
  formatInTimezone,
  formatOrgClock,
  formatOrgClockLong,
  normalizeCountryCode,
  timezoneForPays
} from '../composables/useCountryTimezone'

describe('useCountryTimezone', () => {
  it('maps pays to IANA timezones', () => {
    expect(timezoneForPays('FR')).toBe('Europe/Paris')
    expect(timezoneForPays('BE')).toBe('Europe/Brussels')
    expect(timezoneForPays('MG')).toBe('Indian/Antananarivo')
    expect(timezoneForPays('MA')).toBe('Africa/Casablanca')
    expect(timezoneForPays('TN')).toBe('Africa/Tunis')
    expect(timezoneForPays('CA')).toBe('America/Toronto')
    expect(timezoneForPays('')).toBe('Europe/Paris')
    expect(timezoneForPays('md')).toBe('Indian/Antananarivo')
  })

  it('normalizes country codes', () => {
    expect(normalizeCountryCode('be')).toBe('BE')
    expect(normalizeCountryCode('')).toBe('FR')
    expect(normalizeCountryCode('DE')).toBe('FR')
  })

  it('formats a single timezone clock', () => {
    const iso = '2024-01-15T12:00:00.000Z'
    expect(formatInTimezone(iso, 'Europe/Paris', 'fr', 'time')).toMatch(/13:00/)
    expect(formatInTimezone(iso, 'Indian/Antananarivo', 'fr', 'time')).toMatch(/15:00/)
  })

  it('formatOrgClock stays single-zone', () => {
    const iso = '2024-01-15T12:00:00.000Z'
    expect(formatOrgClock(iso, 'MG', 'fr', 'time')).toMatch(/15:00/)
    expect(formatOrgClock(iso, 'MG', 'fr', 'time')).not.toContain('·')
    expect(formatOrgClock(null, 'FR', 'fr')).toBe('—')
  })

  it('formatOrgClockLong keeps weekday/month', () => {
    const label = formatOrgClockLong(new Date('2024-01-15T12:00:00.000Z'), 'FR', 'fr')
    expect(label.toLowerCase()).toMatch(/janvier|lundi|mardi/)
  })

  it('shows dual clocks with exact wall times when org and client differ', () => {
    const iso = '2024-01-15T12:00:00.000Z'
    expect(formatDualClock(iso, 'MG', 'FR', 'fr', 'time')).toBe('15:00 (MG) · 13:00 (FR)')
  })

  it('keeps a single clock when timezones match or wall clocks equal', () => {
    const iso = '2024-01-15T12:00:00.000Z'
    expect(formatDualClock(iso, 'FR', 'FR', 'fr', 'time')).toBe('13:00')
    expect(formatDualClock(iso, 'FR', 'BE', 'fr', 'time')).toBe('13:00')
  })
})
