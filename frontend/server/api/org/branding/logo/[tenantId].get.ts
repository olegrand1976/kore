export default defineEventHandler(async (event) => {
  const tenantId = getRouterParam(event, 'tenantId')
  const headers = apiAuthHeaders(event)
  return proxyRequest(event, `${apiBase()}/api/v1/branding/logo/${tenantId}`, {
    headers,
    fetchOptions: { redirect: 'manual', responseType: 'arrayBuffer' }
  })
})
