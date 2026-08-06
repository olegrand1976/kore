export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const query = getQuery(event)
  const active = typeof query.active === 'string' ? query.active : undefined
  const qs = active ? `?active=${encodeURIComponent(active)}` : ''
  return $fetch(`${apiBase()}/api/v1/applications${qs}`, { headers })
})
