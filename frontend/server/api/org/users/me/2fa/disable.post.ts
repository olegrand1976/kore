export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  return $fetch(`${apiBase()}/api/v1/users/me/2fa/disable`, {
    method: 'POST',
    headers: apiAuthHeaders(event),
    body
  })
})
