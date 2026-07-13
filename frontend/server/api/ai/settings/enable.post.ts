export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const body = await readBody(event)
  await $fetch(`${apiBase()}/api/v1/ai/settings/enable`, {
    method: 'POST',
    headers: { ...headers, 'Content-Type': 'application/json' },
    body
  })
  setResponseStatus(event, 204)
  return null
})
