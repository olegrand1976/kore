export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const headers = apiAuthHeaders(event)
  return $fetch(`${apiBase()}/api/v1/planning`, { headers, query })
})
