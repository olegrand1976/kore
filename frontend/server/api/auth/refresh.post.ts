export default defineEventHandler(async (event) => {
  const refreshToken = getCookie(event, 'kore_refresh_token')
  if (!refreshToken) {
    throw createError({ statusCode: 401, statusMessage: 'No refresh token' })
  }

  const config = useRuntimeConfig()
  const response = await $fetch(`${config.public.apiBase}/api/v1/auth/refresh`, {
    method: 'POST',
    body: { refreshToken }
  })

  const data = (response as { data?: AuthTokenPayload }).data
  setAuthCookies(event, extractAuthTokens(data))
  return response
})
