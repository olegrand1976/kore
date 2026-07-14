import { createError } from 'h3'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const headers = apiAuthHeaders(event)
  try {
    const response = await $fetch.raw(`${apiBase()}/api/v1/timesheets/${id}/pdf`, {
      method: 'POST',
      headers,
      responseType: 'arrayBuffer'
    })

    const contentType = response.headers.get('content-type') || 'application/pdf'
    const disposition = response.headers.get('content-disposition')
    setHeader(event, 'content-type', contentType)
    if (disposition) {
      setHeader(event, 'content-disposition', disposition)
    }
    return response._data
  } catch (e: unknown) {
    const err = e as {
      statusCode?: number
      statusMessage?: string
      data?: { error?: { message?: string; code?: string } }
    }
    throw createError({
      statusCode: err.statusCode || 500,
      statusMessage: err.statusMessage || 'PDF generation failed',
      data: err.data
    })
  }
})
