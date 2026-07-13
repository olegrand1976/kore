export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const headers = apiAuthHeaders(event)
  return $fetch(`${apiBase()}/api/v1/demands/${id}/validate-creation`, { method: 'POST', headers })
})
