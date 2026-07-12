export default defineEventHandler(async (event) => {
  const tenantId = getRouterParam(event, 'tenantId')
  const headers = apiAuthHeaders(event)
  const response = await $fetch.raw(`${apiBase()}/api/v1/branding/logo/${tenantId}`, {
    headers,
    responseType: 'arrayBuffer'
  })

  const contentType = response.headers.get('content-type') || 'image/png'
  setHeader(event, 'content-type', contentType)
  setHeader(event, 'cache-control', 'private, max-age=3600')
  return response._data
})
