export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const id = getRouterParam(event, 'id')
  return $fetch(`${apiBase()}/api/v1/integrations/taiga/demands/${id}/push`, {
    method: 'POST',
    headers: { ...headers, 'Content-Type': 'application/json' }
  })
})
