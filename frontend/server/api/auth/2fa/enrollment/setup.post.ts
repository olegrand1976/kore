export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  return $fetch(`${apiBase()}/api/v1/auth/2fa/enrollment/setup`, {
    method: 'POST',
    body
  })
})
