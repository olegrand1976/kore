export default defineEventHandler(async (event) => {
  const token = getRouterParam(event, 'token')
  return $fetch(`${apiBase()}/api/v1/public/proforma/${encodeURIComponent(token || '')}/validate`, {
    method: 'POST'
  })
})
