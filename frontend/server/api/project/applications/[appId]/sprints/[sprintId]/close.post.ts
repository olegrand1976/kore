export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const appId = getRouterParam(event, 'appId')
  const sprintId = getRouterParam(event, 'sprintId')
  return $fetch(`${apiBase()}/api/v1/applications/${appId}/sprints/${sprintId}/close`, {
    method: 'POST',
    headers: { ...headers, 'Content-Type': 'application/json' }
  })
})
