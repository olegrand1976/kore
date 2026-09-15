export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const id = getRouterParam(event, 'id')
  return await $fetch(`${apiBase()}/api/v1/request-attachments/${id}`, {
    method: 'DELETE',
    headers
  }).catch((err) => {
    if (err?.statusCode === 204 || err?.response?.status === 204) {
      return null
    }
    throw err
  })
})
