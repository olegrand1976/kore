export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  return $fetch(`${apiBase()}/api/v1/integrations/taiga/sync`, {
    method: 'POST',
    headers: { ...headers, 'Content-Type': 'application/json' }
  })
})
