export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const id = getRouterParam(event, 'id')
  return $fetch(`${apiBase()}/api/v1/ai/explain/${id}`, { headers })
})
