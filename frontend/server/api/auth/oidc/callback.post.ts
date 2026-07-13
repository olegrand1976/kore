export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  const response = await $fetch(`${apiBase()}/api/v1/auth/oidc/callback`, {
    method: 'POST',
    body
  })
  const data = (response as { data?: AuthTokenPayload }).data
  setAuthCookies(event, extractAuthTokens(data))
  return response
})
