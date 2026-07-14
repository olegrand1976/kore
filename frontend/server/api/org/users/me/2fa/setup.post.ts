export default defineEventHandler(async (event) => {
  return $fetch(`${apiBase()}/api/v1/users/me/2fa/setup`, {
    method: 'POST',
    headers: apiAuthHeaders(event)
  })
})
