export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const appId = getRouterParam(event, 'appId')
  const sprintId = getRouterParam(event, 'sprintId')
  const body = await readBody(event)
  return $fetch(`${apiBase()}/api/v1/applications/${appId}/sprints/${sprintId}/plan`, {
    method: 'POST',
    headers: { ...headers, 'Content-Type': 'application/json' },
    body
  })
})
