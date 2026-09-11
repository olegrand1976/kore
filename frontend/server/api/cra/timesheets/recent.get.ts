export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const query = getQuery(event)
  const params = new URLSearchParams()
  for (const key of ['limit', 'year', 'month', 'userId'] as const) {
    const value = query[key]
    if (value == null || value === '') continue
    params.set(key, String(Array.isArray(value) ? value[0] : value))
  }
  const qs = params.toString()
  return $fetch(`${apiBase()}/api/v1/timesheets/recent${qs ? `?${qs}` : ''}`, { headers })
})
