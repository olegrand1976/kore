import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

describe('password-reset BFF', () => {
  it('relays request with public base URL header', () => {
    const src = readFileSync(
      resolve(__dirname, '../../server/api/auth/password-reset/request.post.ts'),
      'utf8'
    )
    expect(src).toContain('/api/v1/auth/password-reset/request')
    expect(src).toContain("method: 'POST'")
  })

  it('relays confirm body to Go confirm endpoint', () => {
    const src = readFileSync(
      resolve(__dirname, '../../server/api/auth/password-reset/confirm.post.ts'),
      'utf8'
    )
    expect(src).toContain('/api/v1/auth/password-reset/confirm')
    expect(src).toContain("method: 'POST'")
  })
})
