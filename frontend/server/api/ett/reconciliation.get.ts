export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const query = getQuery(event)
  const params = new URLSearchParams()
  if (query.month) params.set('month', String(query.month))
  if (query.userId) params.set('userId', String(query.userId))
  return $fetch(`${apiBase()}/api/v1/ett/reconciliation?${params.toString()}`, { headers })
})
