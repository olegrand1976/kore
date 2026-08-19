import { createError } from 'h3'

export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const body = await readBody(event)
  try {
    return await $fetch(`${apiBase()}/api/v1/demands`, {
      method: 'POST',
      headers: { ...headers, 'Content-Type': 'application/json' },
      body
    })
  } catch (e: unknown) {
    const err = e as {
      statusCode?: number
      statusMessage?: string
      data?: { error?: string | { message?: string }; message?: string }
    }
    const nested = err.data?.error
    const message =
      (typeof nested === 'string' ? nested : nested?.message) ||
      err.data?.message ||
      err.statusMessage ||
      'demand create failed'
    throw createError({
      statusCode: err.statusCode || 500,
      statusMessage: message,
      data: err.data
    })
  }
})
