export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  return $fetch(`${apiBase()}/api/v1/ai/public/lead-scoring`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body
  })
})
