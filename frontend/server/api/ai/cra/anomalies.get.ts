export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const query = getQuery(event)
  try {
    return await $fetch(`${apiBase()}/api/v1/ai/cra/anomalies`, { headers, query })
  } catch (err: unknown) {
    // Optional CRA hints — must not block core timesheet flows when AI is disabled.
    if (isAiOptionalUnavailableError(err)) {
      return { data: [] }
    }
    throw err
  }
})
