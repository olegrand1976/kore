export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const headers = apiAuthHeaders(event)
  return sendProxy(event, `${apiBase()}/api/v1/request-attachments/${id}/download`, {
    headers,
    fetchOptions: { redirect: 'manual' }
  })
})
