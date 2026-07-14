export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const query = getQuery(event)
  const month = query.month as string
  const response = await $fetch.raw(`${apiBase()}/api/v1/prestations/export.xml?month=${encodeURIComponent(month)}`, { headers })
  setHeader(event, 'Content-Type', 'application/xml')
  setHeader(event, 'Content-Disposition', 'attachment; filename="prestations-export.xml"')
  return response._data
})
