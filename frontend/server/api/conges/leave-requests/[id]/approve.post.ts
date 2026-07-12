export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const headers = apiAuthHeaders(event)
  return $fetch(`${apiBase()}/api/v1/leave-requests/${id}/approve`, { method: 'POST', headers })
})
