export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  return $fetch(`${apiBase()}/api/v1/notification-rules`, { headers })
})
