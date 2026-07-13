export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const proto = getRequestHeader(event, 'x-forwarded-proto') ?? 'http'
  const host = getRequestHeader(event, 'host') ?? 'localhost'
  const publicBaseUrl = `${proto}://${host}`
  return $fetch(`${apiBase()}/api/v1/public/invitations/resolve`, { query, headers: { 'x-public-base-url': publicBaseUrl } })
})

