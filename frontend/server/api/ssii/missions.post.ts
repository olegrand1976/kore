export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  return $fetch(`${apiBase()}/api/v1/missions`, {
    method: 'POST',
    headers,
    body: await readBody(event)
  })
})
