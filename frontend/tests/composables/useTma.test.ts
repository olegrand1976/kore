import { beforeAll, describe, expect, it, vi } from 'vitest'
import { useTma } from '~/composables/useTma'

const apiFetch = vi.fn()

beforeAll(() => {
  vi.stubGlobal('useApiFetch', () => ({ apiFetch }))
})

describe('useTma pickDescription', () => {
  it('prefers camelCase over PascalCase', () => {
    const { pickDescription } = useTma()
    expect(pickDescription({ description: 'initiale', Description: 'other' })).toBe('initiale')
  })

  it('falls back to PascalCase then empty string', () => {
    const { pickDescription } = useTma()
    expect(pickDescription({ Description: 'from-api' })).toBe('from-api')
    expect(pickDescription({})).toBe('')
  })
})

describe('useTma list field pickers', () => {
  it('reads application, assignee, priority and due date with camelCase preference', () => {
    const { pickApplicationId, pickAssigneeId, pickPriority, pickDueAt } = useTma()
    expect(
      pickApplicationId({ applicationId: 'app-1', ApplicationID: 'app-other' })
    ).toBe('app-1')
    expect(pickApplicationId({ ApplicationID: 'app-2' })).toBe('app-2')
    expect(pickAssigneeId({ assigneeId: 'user-1', AssigneeID: 'user-other' })).toBe('user-1')
    expect(pickAssigneeId({})).toBe('')
    expect(pickPriority({ priority: 'high', Priority: 'low' })).toBe('high')
    expect(pickPriority({})).toBe('normal')
    expect(pickDueAt({ dueAt: '2026-01-01', DueAt: '2025-01-01' })).toBe('2026-01-01')
    expect(pickDueAt({})).toBe('')
  })
})

describe('useTma remove', () => {
  it('calls DELETE on the demand BFF route', async () => {
    apiFetch.mockResolvedValueOnce({ data: { status: 'deleted' } })
    const { remove } = useTma()
    await remove('demand-1')
    expect(apiFetch).toHaveBeenCalledWith('/api/tma/demands/demand-1', { method: 'DELETE' })
  })
})
