export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const query = getQuery(event)
  return $fetch(`${apiBase()}/api/v1/leave-type-configs`, { headers, query })
})
