export default defineEventHandler(async (event) => {
  const headers = await bffAuthHeaders(event)
  const body = await readBody(event)
  return $fetch(`${apiBase()}/api/v1/ai/mobile/voice-cra`, {
    method: 'POST',
    headers,
    body
  })
})
