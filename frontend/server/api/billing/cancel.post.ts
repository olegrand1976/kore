export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  return $fetch(`${apiBase()}/api/v1/billing/cancel`, { method: 'POST', headers })
})
