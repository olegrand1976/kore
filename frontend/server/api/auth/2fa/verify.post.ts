export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const body = await readBody(event)
  const response = await $fetch(`${config.public.apiBase}/api/v1/auth/2fa/verify`, {
    method: 'POST',
    body
  })
  const data = (response as { data?: AuthTokenPayload }).data
  setAuthCookies(event, extractAuthTokens(data))
  return response
})
