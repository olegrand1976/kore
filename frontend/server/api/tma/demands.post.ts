export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const body = await readBody(event)
  return $fetch(`${apiBase()}/api/v1/demands`, {
    method: 'POST',
    headers: { ...headers, 'Content-Type': 'application/json' },
    body
  })
})
