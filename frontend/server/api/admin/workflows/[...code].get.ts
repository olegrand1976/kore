import { createError } from 'h3'

export default defineEventHandler(async (event) => {
  const raw = getRouterParam(event, 'code')
  const code = Array.isArray(raw) ? raw.join('/') : raw
  if (!code) {
    throw createError({ statusCode: 400, statusMessage: 'code required' })
  }

  const headers = apiAuthHeaders(event)
  try {
    return await $fetch(`${apiBase()}/api/v1/workflows/${encodeURIComponent(code)}`, { headers })
  } catch (e: unknown) {
    const err = e as {
      statusCode?: number
      statusMessage?: string
      data?: { error?: { message?: string } }
    }
    throw createError({
      statusCode: err.statusCode || 500,
      statusMessage: err.data?.error?.message || err.statusMessage || 'workflow fetch failed',
      data: err.data
    })
  }
})
