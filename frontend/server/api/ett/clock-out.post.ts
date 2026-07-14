export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  return $fetch(`${apiBase()}/api/v1/ett/clock-out`, {
    method: 'POST',
    headers,
    body: await readBody(event)
  })
})
