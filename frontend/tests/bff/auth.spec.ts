import { describe, expect, it } from 'vitest'
import {
  authorizationHeadersFromToken,
  extractAuthTokens,
  isPartialAuth,
  parseSessionFromToken
} from '../../server/utils/auth'

function fakeJwt(payload: Record<string, unknown>): string {
  const header = Buffer.from(JSON.stringify({ alg: 'none', typ: 'JWT' })).toString('base64url')
  const body = Buffer.from(JSON.stringify(payload)).toString('base64url')
  return `${header}.${body}.sig`
}

describe('BFF auth utils', () => {
  describe('extractAuthTokens', () => {
    it('reads camelCase tokens', () => {
      expect(extractAuthTokens({ accessToken: 'a', refreshToken: 'r' })).toEqual({
        accessToken: 'a',
        refreshToken: 'r'
      })
    })

    it('reads PascalCase tokens', () => {
      expect(extractAuthTokens({ AccessToken: 'A', RefreshToken: 'R' })).toEqual({
        accessToken: 'A',
        refreshToken: 'R'
      })
    })

    it('returns undefined when data missing', () => {
      expect(extractAuthTokens(undefined)).toEqual({
        accessToken: undefined,
        refreshToken: undefined
      })
    })
  })

  describe('isPartialAuth', () => {
    it('detects 2FA challenge', () => {
      expect(isPartialAuth({ requires2FA: true })).toBe(true)
      expect(isPartialAuth({ Requires2FAEnrollment: true })).toBe(true)
      expect(isPartialAuth({ accessToken: 'x' })).toBe(false)
      expect(isPartialAuth(undefined)).toBe(false)
    })
  })

  describe('authorizationHeadersFromToken', () => {
    it('returns empty object without token', () => {
      expect(authorizationHeadersFromToken(undefined)).toEqual({})
      expect(authorizationHeadersFromToken(null)).toEqual({})
      expect(authorizationHeadersFromToken('')).toEqual({})
    })

    it('maps cookie value to Bearer header', () => {
      expect(authorizationHeadersFromToken('tok-123')).toEqual({
        Authorization: 'Bearer tok-123'
      })
    })
  })

  describe('parseSessionFromToken', () => {
    it('returns null for invalid tokens', () => {
      expect(parseSessionFromToken(undefined)).toBeNull()
      expect(parseSessionFromToken('not-a-jwt')).toBeNull()
      expect(parseSessionFromToken('a.b')).toBeNull()
    })

    it('extracts profile, sub and tenant_id', () => {
      const token = fakeJwt({
        sub: 'user-1',
        tenant_id: 'tenant-9',
        profile: 'Administrateur',
        roles: ['admin']
      })
      expect(parseSessionFromToken(token)).toEqual({
        profile: 'Administrateur',
        userId: 'user-1',
        tenantId: 'tenant-9',
        roles: ['admin']
      })
    })
  })
})
