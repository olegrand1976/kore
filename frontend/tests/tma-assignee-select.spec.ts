import { describe, expect, it } from 'vitest'
import { syncAssigneeFromDemand, syncAssigneeFromUsers } from '../utils/tmaAssigneeSelect'

describe('syncAssigneeFromDemand', () => {
  it('prefers the demand assignee when present', () => {
    expect(syncAssigneeFromDemand('user-b', 'user-a')).toBe('user-b')
    expect(syncAssigneeFromDemand('user-b', '')).toBe('user-b')
  })

  it('keeps the current selection when the demand has no assignee', () => {
    expect(syncAssigneeFromDemand(undefined, 'user-a')).toBe('user-a')
    expect(syncAssigneeFromDemand(null, 'user-a')).toBe('user-a')
    expect(syncAssigneeFromDemand('', 'user-a')).toBe('user-a')
  })
})

describe('syncAssigneeFromUsers', () => {
  const users = [{ id: 'u1' }, { id: 'u2' }]

  it('defaults to the first user only when empty', () => {
    expect(syncAssigneeFromUsers('', users)).toBe('u1')
    expect(syncAssigneeFromUsers('', null)).toBe('')
    expect(syncAssigneeFromUsers('', [])).toBe('')
  })

  it('does not overwrite an in-progress selection when users refresh', () => {
    expect(syncAssigneeFromUsers('u2', users)).toBe('u2')
    expect(syncAssigneeFromUsers('u2', [{ id: 'u1' }, { id: 'u2' }, { id: 'u3' }])).toBe('u2')
  })
})
