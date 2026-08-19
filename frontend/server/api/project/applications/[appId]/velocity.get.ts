export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const appId = getRouterParam(event, 'appId')
  return $fetch(`${apiBase()}/api/v1/applications/${appId}/velocity`, { headers })
})
