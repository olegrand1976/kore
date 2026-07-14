export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const code = getRouterParam(event, 'code')
  return $fetch(`${apiBase()}/api/v1/dashboards/${code}`, { headers })
})
