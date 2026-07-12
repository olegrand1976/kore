export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  try {
    await $fetch(`${apiBase()}/api/v1/auth/logout`, { method: 'POST', headers })
  } catch {
    // best-effort server-side logout
  }
  deleteCookie(event, 'kore_access_token')
  deleteCookie(event, 'kore_refresh_token')
  return { data: { status: 'logged_out' } }
})
