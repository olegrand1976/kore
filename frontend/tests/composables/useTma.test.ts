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

describe('useTma remove', () => {
  it('calls DELETE on the demand BFF route', async () => {
    apiFetch.mockResolvedValueOnce({ data: { status: 'deleted' } })
    const { remove } = useTma()
    await remove('demand-1')
    expect(apiFetch).toHaveBeenCalledWith('/api/tma/demands/demand-1', { method: 'DELETE' })
  })
})
