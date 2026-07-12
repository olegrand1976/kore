export default defineEventHandler((event) => {
  deleteCookie(event, 'kore_access_token')
  deleteCookie(event, 'kore_refresh_token')
  return { data: { status: 'logged_out' } }
})
