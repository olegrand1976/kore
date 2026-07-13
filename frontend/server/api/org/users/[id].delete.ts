export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const id = getRouterParam(event, 'id')
  return $fetch(`${apiBase()}/api/v1/users/${id}`, {
    method: 'DELETE',
    headers
  })
})
