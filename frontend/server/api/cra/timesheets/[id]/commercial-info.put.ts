export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const headers = apiAuthHeaders(event)
  const body = await readBody(event)
  return $fetch(`${apiBase()}/api/v1/timesheets/${id}/commercial-info`, {
    method: 'PUT',
    headers: { ...headers, 'Content-Type': 'application/json' },
    body
  })
})
