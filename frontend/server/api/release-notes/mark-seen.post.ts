import { apiAuthHeaders, apiBase } from '../../utils/auth'

function currentVersion(): string {
  return process.env.KORE_VERSION || process.env.NUXT_PUBLIC_KORE_VERSION || process.env.APP_VERSION || '0.0.0'
}

export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  if (!headers.Authorization) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthorized' })
  }

  const version = currentVersion()
  await $fetch(`${apiBase()}/api/v1/users/me/release-notes/seen`, {
    method: 'POST',
    headers,
    body: { version }
  })
  return { status: 'ok' }
})

