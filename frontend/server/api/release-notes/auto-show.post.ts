import { apiAuthHeaders, apiBase } from '../../utils/auth'

export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  if (!headers.Authorization) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthorized' })
  }

  const body = await readBody<{ enabled?: boolean }>(event)
  const enabled = body?.enabled === true

  await $fetch(`${apiBase()}/api/v1/users/me/release-notes/auto-show`, {
    method: 'POST',
    headers,
    body: { enabled }
  })
  return { status: 'ok' }
})

