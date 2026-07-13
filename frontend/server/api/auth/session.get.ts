export default defineEventHandler((event) => {
  const token = getCookie(event, 'kore_access_token')
  if (!token) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthorized' })
  }

  const parts = token.split('.')
  if (parts.length !== 3) {
    throw createError({ statusCode: 401, statusMessage: 'Invalid token' })
  }

  try {
    const payload = JSON.parse(Buffer.from(parts[1], 'base64url').toString('utf8'))
    const roles = Array.isArray(payload.roles) ? (payload.roles as string[]) : []
    return {
      ok: true,
      profile: payload.profile as string | undefined,
      userId: payload.sub as string | undefined,
      tenantId: payload.tenant_id as string | undefined,
      isPlatformAdmin: roles.includes('platform_admin')
    }
  } catch {
    throw createError({ statusCode: 401, statusMessage: 'Invalid token' })
  }
})
