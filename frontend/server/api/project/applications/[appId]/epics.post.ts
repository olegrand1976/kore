export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const appId = getRouterParam(event, 'appId')
  const body = await readBody(event)
  return $fetch(`${apiBase()}/api/v1/applications/${appId}/epics`, {
    method: 'POST',
    headers: { ...headers, 'Content-Type': 'application/json' },
    body
  })
})
