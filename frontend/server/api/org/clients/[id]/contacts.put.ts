export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const id = getRouterParam(event, 'id')
  const body = await readBody(event)
  return $fetch(`${apiBase()}/api/v1/clients/${id}/contacts`, {
    method: 'PUT',
    headers: { ...headers, 'Content-Type': 'application/json' },
    body
  })
})
