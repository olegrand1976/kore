export default defineEventHandler(async (event) => {
  const code = getRouterParam(event, 'code')
  const headers = apiAuthHeaders(event)
  return $fetch(`${apiBase()}/api/v1/workflows/${code}`, { headers })
})
