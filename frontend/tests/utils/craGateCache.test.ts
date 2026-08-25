import { describe, expect, it } from 'vitest'
import {
  CRA_GATE_TTL_MS,
  readCachedCraGateMode,
  writeCraGateModeCache
} from '~/utils/craGateCache'

describe('craGateCache', () => {
  it('returns null when cache is empty or expired', () => {
    expect(readCachedCraGateMode(null)).toBeNull()
    const entry = writeCraGateModeCache('warn', 1_000)
    expect(readCachedCraGateMode(entry, 1_000 + CRA_GATE_TTL_MS)).toBeNull()
  })

  it('returns mode within TTL', () => {
    const entry = writeCraGateModeCache('block', 5_000)
    expect(readCachedCraGateMode(entry, 5_000 + CRA_GATE_TTL_MS - 1)).toBe('block')
  })
})
