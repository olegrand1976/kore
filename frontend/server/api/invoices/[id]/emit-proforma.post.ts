export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const id = getRouterParam(event, 'id')
  const body = await readBody(event).catch(() => ({}))
  const proto = getRequestHeader(event, 'x-forwarded-proto') || 'http'
  const host = getRequestHeader(event, 'host') || 'localhost'
  return $fetch(`${apiBase()}/api/v1/invoices/${id}/emit-proforma`, {
    method: 'POST',
    headers: {
      ...headers,
      'x-public-base-url': `${proto}://${host}`
    },
    body
  })
})
