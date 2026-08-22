export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const headers = apiAuthHeaders(event)
  return $fetch(`${apiBase()}/api/v1/integrations/taiga/links/by-application/${id}`, { headers })
})
