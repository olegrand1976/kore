export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const id = getRouterParam(event, 'id')
  return $fetch(`${apiBase()}/api/v1/missions/${id}/applications`, {
    method: 'PUT',
    headers,
    body: await readBody(event)
  })
})
