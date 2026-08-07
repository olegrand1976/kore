export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const id = getRouterParam(event, 'id')
  const body = await readBody(event).catch(() => ({}))
  return $fetch(`${apiBase()}/api/v1/invoices/${id}/emit-proforma`, {
    method: 'POST',
    headers,
    body
  })
})
