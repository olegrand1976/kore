export default defineEventHandler(async (event) => {
  const limit = getQuery(event).limit
  const headers = apiAuthHeaders(event)
  const query = limit ? `?limit=${limit}` : ''
  return $fetch(`${apiBase()}/api/v1/timesheets/recent${query}`, { headers })
})
