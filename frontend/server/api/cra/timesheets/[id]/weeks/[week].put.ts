export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const week = getRouterParam(event, 'week')
  const headers = apiAuthHeaders(event)
  const body = await readBody(event)
  return $fetch(`${apiBase()}/api/v1/timesheets/${id}/weeks/${week}`, {
    method: 'PUT',
    headers: { ...headers, 'Content-Type': 'application/json' },
    body
  })
})
