import { describe, expect, it } from 'vitest'
import { pickDemandId, pickEpicId, pickSprintId } from '~/composables/useProject'

describe('useProject pick helpers', () => {
  it('pickEpicId prefers camelCase', () => {
    expect(pickEpicId({ id: 'a', ID: 'b' })).toBe('a')
    expect(pickEpicId({ ID: 'b' })).toBe('b')
  })

  it('pickSprintId and pickDemandId mirror epic helper', () => {
    expect(pickSprintId({ id: 's1' })).toBe('s1')
    expect(pickDemandId({ demandId: 'd1' })).toBe('d1')
    expect(pickDemandId({ DemandID: 'd2' })).toBe('d2')
  })
})
