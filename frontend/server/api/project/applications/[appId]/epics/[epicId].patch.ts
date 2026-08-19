export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const appId = getRouterParam(event, 'appId')
  const epicId = getRouterParam(event, 'epicId')
  const body = await readBody(event)
  return $fetch(`${apiBase()}/api/v1/applications/${appId}/epics/${epicId}`, {
    method: 'PATCH',
    headers,
    body
  })
})
