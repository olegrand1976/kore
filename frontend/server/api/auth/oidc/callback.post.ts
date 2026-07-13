import { createError } from 'h3'

export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  const apiUrl = `${apiBase()}/api/v1/auth/oidc/callback`

  try {
    const response = await $fetch(apiUrl, { method: 'POST', body })
    const data = (response as { data?: AuthTokenPayload }).data
    setAuthCookies(event, extractAuthTokens(data))
    return response
  } catch (e: any) {
    throw createError({
      statusCode: e?.statusCode || 500,
      statusMessage: e?.statusMessage,
      data: e?.data
    })
  }
})
