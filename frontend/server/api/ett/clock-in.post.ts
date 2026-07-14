export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  return $fetch(`${apiBase()}/api/v1/ett/clock-in`, {
    method: 'POST',
    headers,
    body: await readBody(event)
  })
})
