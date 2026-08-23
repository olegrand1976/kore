import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'
import { isAiOptionalUnavailableError } from '../../server/utils/aiOptional'

const root = join(__dirname, '../..')

function readRoute(rel: string): string {
  return readFileSync(join(root, 'server/api', rel), 'utf8')
}

describe('BFF AI CRA anomalies route', () => {
  it('proxies anomalies to Go and tolerates AI-disabled responses only', () => {
    const src = readRoute('ai/cra/anomalies.get.ts')
    expect(src).toContain('apiAuthHeaders(event)')
    expect(src).toContain('/api/v1/ai/cra/anomalies')
    expect(src).toContain('isAiOptionalUnavailableError(err)')
    expect(src).toContain('{ data: [] }')
    expect(src).not.toContain('status === 403')
  })
})

describe('isAiOptionalUnavailableError', () => {
  it('returns true for tenant AI disabled', () => {
    expect(
      isAiOptionalUnavailableError({
        statusCode: 403,
        data: { error: { code: 'FORBIDDEN', message: 'ai assistance disabled for tenant' } }
      })
    ).toBe(true)
  })

  it('returns true for capability off', () => {
    expect(
      isAiOptionalUnavailableError({
        statusCode: 403,
        data: { error: { code: 'FORBIDDEN', message: 'ai capability disabled' } }
      })
    ).toBe(true)
  })

  it('returns false for unrelated forbidden errors', () => {
    expect(
      isAiOptionalUnavailableError({
        statusCode: 403,
        data: { error: { code: 'FORBIDDEN', message: 'insufficient permissions' } }
      })
    ).toBe(false)
  })

  it('returns false for non-403 errors', () => {
    expect(
      isAiOptionalUnavailableError({
        statusCode: 500,
        data: { error: { code: 'INTERNAL', message: 'ai assistance disabled for tenant' } }
      })
    ).toBe(false)
  })
})
