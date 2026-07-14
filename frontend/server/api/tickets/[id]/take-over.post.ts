export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const headers = apiAuthHeaders(event)
  return $fetch(`${apiBase()}/api/v1/tickets/${id}/take-over`, { method: 'POST', headers })
})
