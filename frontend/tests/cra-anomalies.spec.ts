import { describe, expect, it } from 'vitest'
import { normalizeAnomalyMessages } from '../utils/craAnomalies'

describe('normalizeAnomalyMessages', () => {
  it('reads anomalies from data array (Go envelope)', () => {
    expect(
      normalizeAnomalyMessages({
        data: [{ code: 'DAY_CAPACITY', message: 'Heures supérieures à 8h sur la journée', day: '2026-08-23' }]
      })
    ).toEqual(['Heures supérieures à 8h sur la journée'])
  })

  it('reads anomalies from legacy data.anomalies shape', () => {
    expect(
      normalizeAnomalyMessages({
        data: { anomalies: ['Semaine 5 sans saisie'] }
      })
    ).toEqual(['Semaine 5 sans saisie'])
  })

  it('falls back to code when message is empty', () => {
    expect(normalizeAnomalyMessages({ data: [{ code: 'EMPTY_WEEK' }] })).toEqual(['EMPTY_WEEK'])
  })

  it('returns empty array for invalid payloads', () => {
    expect(normalizeAnomalyMessages(null)).toEqual([])
    expect(normalizeAnomalyMessages({ data: { anomalies: 'nope' } })).toEqual([])
  })
})
