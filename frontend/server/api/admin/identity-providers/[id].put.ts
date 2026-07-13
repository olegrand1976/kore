export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const headers = apiAuthHeaders(event)
  const body = await readBody(event)
  return $fetch(`${apiBase()}/api/v1/admin/identity-providers/${id}`, {
    method: 'PUT',
    headers,
    body
  })
})
