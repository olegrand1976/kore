export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  return $fetch(`${apiBase()}/api/v1/ai/public/chat`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body
  })
})
