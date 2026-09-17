export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  return $fetch(`${apiBase()}/api/v1/auth/password-reset/request`, {
    method: 'POST',
    body
  })
})
