import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

describe('public signup BFF', () => {
  it('relays POST body to Go /api/v1/public/signup', () => {
    const src = readFileSync(
      resolve(__dirname, '../../server/api/public/signup.post.ts'),
      'utf8'
    )
    expect(src).toContain('/api/v1/public/signup')
    expect(src).toContain("method: 'POST'")
  })
})

describe('platform tenants BFF', () => {
  it('relays authenticated POST to Go /api/v1/platform/tenants', () => {
    const src = readFileSync(
      resolve(__dirname, '../../server/api/platform/tenants.post.ts'),
      'utf8'
    )
    expect(src).toContain('/api/v1/platform/tenants')
    expect(src).toContain('apiAuthHeaders')
  })
})
