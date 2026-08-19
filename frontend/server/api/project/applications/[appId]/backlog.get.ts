export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const appId = getRouterParam(event, 'appId')
  const query = getQuery(event)
  const qs = query.backlogOnly === 'true' ? '?backlogOnly=true' : ''
  return $fetch(`${apiBase()}/api/v1/applications/${appId}/backlog${qs}`, { headers })
})
