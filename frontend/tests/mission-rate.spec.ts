import { describe, expect, it } from 'vitest'

type RateUnit = 'tjm' | 'hourly'

function normalizeRateUnit(value: unknown): RateUnit {
  switch (String(value ?? '').toLowerCase()) {
    case 'hourly':
      return 'hourly'
    case 'tjm':
    default:
      return 'tjm'
  }
}

function formatRateLabel(amountLabel: string, rateUnit: unknown): string {
  switch (normalizeRateUnit(rateUnit)) {
    case 'hourly':
      return `${amountLabel}/h`
    case 'tjm':
      return `${amountLabel}/j`
    default: {
      const _exhaustive: never = normalizeRateUnit(rateUnit)
      return _exhaustive
    }
  }
}

describe('mission rate unit', () => {
  it('normalizes rate units', () => {
    expect(normalizeRateUnit('hourly')).toBe('hourly')
    expect(normalizeRateUnit('TJM')).toBe('tjm')
    expect(normalizeRateUnit('')).toBe('tjm')
    expect(normalizeRateUnit('weekly')).toBe('tjm')
  })

  it('formats rate labels for list/fiche', () => {
    expect(formatRateLabel('500 €', 'tjm')).toBe('500 €/j')
    expect(formatRateLabel('62,50 €', 'hourly')).toBe('62,50 €/h')
  })
})
