export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const month = query.month
  if (!month || Array.isArray(month)) {
    throw createError({ statusCode: 400, statusMessage: 'month required' })
  }
  const headers = apiAuthHeaders(event)
  const params = new URLSearchParams({ month: String(month) })
  const userId = query.userId
  if (userId && !Array.isArray(userId) && String(userId).trim()) {
    params.set('userId', String(userId).trim())
  }
  return $fetch(`${apiBase()}/api/v1/timesheets?${params.toString()}`, { headers })
})
