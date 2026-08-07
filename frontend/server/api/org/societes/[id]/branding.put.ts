export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const headers = apiAuthHeaders(event)
  // Forward raw multipart (Content-Type + boundary) without Blob re-encode.
  // Upstream 4xx/5xx are proxied as-is (sendProxy ignoreResponseError).
  return proxyRequest(event, `${apiBase()}/api/v1/societes/${id}/branding`, {
    headers,
    fetchOptions: { redirect: 'manual' }
  })
})
