export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const query = getQuery(event)
  const qs = new URLSearchParams()
  if (typeof query.applicationId === 'string' && query.applicationId) {
    qs.set('applicationId', query.applicationId)
  }
  const suffix = qs.toString() ? `?${qs.toString()}` : ''
  return $fetch(`${apiBase()}/api/v1/budgets${suffix}`, { headers })
})
