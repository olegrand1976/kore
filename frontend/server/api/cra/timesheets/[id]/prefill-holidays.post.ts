export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const headers = apiAuthHeaders(event)
  const query = getQuery(event)
  const country = query.country ? `?country=${encodeURIComponent(String(query.country))}` : ''
  return $fetch(`${apiBase()}/api/v1/timesheets/${id}/prefill-holidays${country}`, {
    method: 'POST',
    headers
  })
})
