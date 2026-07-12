export default defineEventHandler(async (event) => {
  const month = getQuery(event).month
  if (!month) {
    throw createError({ statusCode: 400, statusMessage: 'month required' })
  }
  const headers = apiAuthHeaders(event)
  return $fetch(`${apiBase()}/api/v1/timesheets?month=${month}`, { headers })
})
