export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const headers = apiAuthHeaders(event)
  const response = await $fetch.raw(`${apiBase()}/api/v1/timesheets/${id}/pdf`, {
    method: 'POST',
    headers,
    responseType: 'arrayBuffer'
  })

  const contentType = response.headers.get('content-type') || 'text/html'
  const disposition = response.headers.get('content-disposition')
  setHeader(event, 'content-type', contentType)
  if (disposition) {
    setHeader(event, 'content-disposition', disposition)
  }
  return response._data
})
