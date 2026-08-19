export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const appId = getRouterParam(event, 'appId')
  const body = await readBody(event)
  return $fetch(`${apiBase()}/api/v1/applications/${appId}/kanban-config`, {
    method: 'PUT',
    headers,
    body
  })
})
