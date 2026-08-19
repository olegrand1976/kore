import { describe, expect, it } from 'vitest'
import { isAgileProfile, METHODOLOGY_PROFILES } from '~/composables/useMethodologyTerms'

describe('useMethodologyTerms helpers', () => {
  it('exposes methodology profiles', () => {
    expect(METHODOLOGY_PROFILES).toContain('agile_scrum')
  })

  it('detects agile profiles', () => {
    expect(isAgileProfile('agile_scrum')).toBe(true)
    expect(isAgileProfile('agile_kanban')).toBe(true)
    expect(isAgileProfile('psa')).toBe(false)
  })
})
