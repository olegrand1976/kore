export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const response = await $fetch.raw(`${apiBase()}/api/v1/demands/export.xml`, { headers })
  setHeader(event, 'Content-Type', 'application/xml')
  setHeader(event, 'Content-Disposition', 'attachment; filename="demands-export.xml"')
  return response._data
})
