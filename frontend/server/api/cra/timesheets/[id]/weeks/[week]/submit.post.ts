export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const week = getRouterParam(event, 'week')
  const headers = apiAuthHeaders(event)
  return $fetch(`${apiBase()}/api/v1/timesheets/${id}/weeks/${week}/submit`, {
    method: 'POST',
    headers
  })
})
