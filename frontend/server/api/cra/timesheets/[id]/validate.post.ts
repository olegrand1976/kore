import { createError } from 'h3'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const headers = apiAuthHeaders(event)
  const body = await readBody<{ force?: boolean }>(event).catch(() => ({}))
  try {
    return await $fetch(`${apiBase()}/api/v1/timesheets/${id}/validate`, {
      method: 'POST',
      headers,
      body: body ?? {}
    })
  } catch (e: unknown) {
    const err = e as {
      statusCode?: number
      statusMessage?: string
      data?: { error?: string | { message?: string; code?: string }; message?: string }
    }
    const nested = err.data?.error
    const message =
      (typeof nested === 'string' ? nested : nested?.message) ||
      err.data?.message ||
      err.statusMessage ||
      'timesheet validate failed'
    throw createError({
      statusCode: err.statusCode || 500,
      statusMessage: message,
      data: err.data
    })
  }
})
