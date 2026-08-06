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

  try {
    return await $fetch(`${apiBase()}/api/v1/societes/${id}/branding`, {
      method: 'PUT',
      headers,
      body
    })
  } catch (err: unknown) {
    const e = err as { statusCode?: number; status?: number; data?: { error?: { message?: string } }; message?: string }
    const status = e.statusCode ?? e.status ?? 500
    const message = e.data?.error?.message ?? e.message ?? 'branding update failed'
    throw createError({ statusCode: status, message })
  }
})
