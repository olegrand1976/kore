type AuthTokenPayload = {
  accessToken?: string
  refreshToken?: string
  AccessToken?: string
  RefreshToken?: string
}

export function extractAuthTokens(data: AuthTokenPayload | undefined) {
  if (!data) {
    return { accessToken: undefined, refreshToken: undefined }
  }
  return {
    accessToken: data.accessToken ?? data.AccessToken,
    refreshToken: data.refreshToken ?? data.RefreshToken
  }
}

function shouldUseSecureCookies(event: import('h3').H3Event): boolean {
  const forwardedProto = getRequestHeader(event, 'x-forwarded-proto')
  if (forwardedProto === 'https') {
    return true
  }
  const host = getRequestHeader(event, 'host') ?? ''
  if (/^(localhost|127\.0\.0\.1)(:\d+)?$/i.test(host)) {
    return false
  }
  return process.env.NODE_ENV === 'production'
}

export function setAuthCookies(
  event: import('h3').H3Event,
  tokens: { accessToken?: string; refreshToken?: string }
) {
  const secure = shouldUseSecureCookies(event)
  if (tokens.accessToken) {
    setCookie(event, 'kore_access_token', tokens.accessToken, {
      httpOnly: true,
      secure,
      sameSite: 'lax',
      path: '/'
    })
  }
  if (tokens.refreshToken) {
    setCookie(event, 'kore_refresh_token', tokens.refreshToken, {
      httpOnly: true,
      secure,
      sameSite: 'lax',
      path: '/'
    })
  }
}

export function apiAuthHeaders(event: Parameters<typeof defineEventHandler>[0] extends never ? never : import('h3').H3Event): Record<string, string> {
  const token = getCookie(event, 'kore_access_token')
  if (!token) {
    return {}
  }
  return { Authorization: `Bearer ${token}` }
}

export function apiBase(): string {
  return useRuntimeConfig().public.apiBase
}

export type SessionPayload = {
  profile?: string
  userId?: string
  tenantId?: string
}

export function parseSessionFromEvent(event: import('h3').H3Event): SessionPayload | null {
  const token = getCookie(event, 'kore_access_token')
  if (!token) return null

  const parts = token.split('.')
  if (parts.length !== 3) return null

  try {
    const payload = JSON.parse(Buffer.from(parts[1], 'base64url').toString('utf8'))
    return {
      profile: payload.profile as string | undefined,
      userId: payload.sub as string | undefined,
      tenantId: payload.tenant_id as string | undefined
    }
  } catch {
    return null
  }
}
