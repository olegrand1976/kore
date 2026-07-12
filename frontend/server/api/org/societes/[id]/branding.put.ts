export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const headers = apiAuthHeaders(event)
  const form = await readMultipartFormData(event)
  if (!form) {
    throw createError({ statusCode: 400, message: 'invalid form' })
  }

  const body = new FormData()
  for (const part of form) {
    if (!part.name) continue
    if (part.filename) {
      body.append(part.name, new Blob([part.data], { type: part.type }), part.filename)
    } else {
      body.append(part.name, part.data.toString())
    }
  }

  return $fetch(`${apiBase()}/api/v1/societes/${id}/branding`, {
    method: 'PUT',
    headers,
    body
  })
})
