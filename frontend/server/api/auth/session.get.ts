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
    return {
      ok: true,
      profile: payload.profile as string | undefined,
      userId: payload.sub as string | undefined,
      tenantId: payload.tenant_id as string | undefined
    }
  } catch {
    throw createError({ statusCode: 401, statusMessage: 'Invalid token' })
  }
})
