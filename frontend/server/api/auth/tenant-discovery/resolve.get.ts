export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  return $fetch(`${apiBase()}/api/v1/auth/tenant-discovery/resolve`, { query })
})

