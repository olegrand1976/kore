import { apiAuthHeaders, apiBase } from '../../utils/auth'

type BackendPrefs = {
  lastSeenVersion: string | null
  autoShowEnabled: boolean
}

function currentVersion(): string {
  return process.env.KORE_VERSION || process.env.NUXT_PUBLIC_KORE_VERSION || process.env.APP_VERSION || '0.0.0'
}

export default defineEventHandler(async (event) => {
  const headers = apiAuthHeaders(event)
  const version = currentVersion()

  if (!headers.Authorization) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthorized' })
  }

  try {
    const prefs = await $fetch<BackendPrefs>(`${apiBase()}/api/v1/users/me/release-notes`, { headers })
    return { currentVersion: version, lastSeenVersion: prefs.lastSeenVersion, autoShowEnabled: prefs.autoShowEnabled }
  } catch {
    // Non-bloquant: si le backend n'est pas encore à jour, on renvoie des defaults.
    return { currentVersion: version, lastSeenVersion: null, autoShowEnabled: true }
  }
})

