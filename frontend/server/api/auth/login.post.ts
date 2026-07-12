export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const body = await readBody(event)
  const response = await $fetch(`${config.public.apiBase}/api/v1/auth/login`, {
    method: 'POST',
    body
  })
  const data = (response as any).data
  if (data?.accessToken) {
    setCookie(event, 'kore_access_token', data.accessToken, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      path: '/'
    })
    if (data.refreshToken) {
      setCookie(event, 'kore_refresh_token', data.refreshToken, {
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        sameSite: 'lax',
        path: '/'
      })
    }
  }
  return response
})
