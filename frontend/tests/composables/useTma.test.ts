import { beforeAll, describe, expect, it, vi } from 'vitest'
import { useTma } from '~/composables/useTma'

beforeAll(() => {
  vi.stubGlobal('useApiFetch', () => ({ apiFetch: vi.fn() }))
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
