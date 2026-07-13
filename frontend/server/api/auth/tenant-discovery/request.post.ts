export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  const proto = getRequestHeader(event, 'x-forwarded-proto') ?? 'http'
  const host = getRequestHeader(event, 'host') ?? 'localhost'
  const publicBaseUrl = `${proto}://${host}`
  return $fetch(`${apiBase()}/api/v1/auth/tenant-discovery/request`, {
    method: 'POST',
    body,
    headers: { 'x-public-base-url': publicBaseUrl }
  })
})

