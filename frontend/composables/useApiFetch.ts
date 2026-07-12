type FetchOptions = Record<string, unknown>
type FetchFn = <T>(url: string, options?: FetchOptions) => Promise<T>

function statusOf(err: unknown): number | undefined {
  const e = err as { statusCode?: number; status?: number; response?: { status?: number } }
  return e?.statusCode ?? e?.status ?? e?.response?.status
}

// Core retry logic, isolated from Nuxt globals so it can be unit-tested.
// On a 401, attempts a single refresh then replays the request once.
export async function fetchWithRefresh<T>(
  fetchFn: FetchFn,
  refreshFn: () => Promise<boolean>,
  onAuthFailure: () => void | Promise<void>,
  url: string,
  options?: FetchOptions
): Promise<T> {
  try {
    return await fetchFn<T>(url, options)
  } catch (err) {
    if (statusOf(err) !== 401) {
      throw err
    }
    const refreshed = await refreshFn()
    if (!refreshed) {
      await onAuthFailure()
      throw err
    }
    return await fetchFn<T>(url, options)
  }
}

let refreshInFlight: Promise<boolean> | null = null

export function useApiFetch() {
  const refresh = async (): Promise<boolean> => {
    if (!refreshInFlight) {
      refreshInFlight = $fetch('/api/auth/refresh', { method: 'POST' })
        .then(() => true)
        .catch(() => false)
        .finally(() => {
          refreshInFlight = null
        })
    }
    return refreshInFlight
  }

  const onAuthFailure = async () => {
    if (import.meta.client) {
      await navigateTo('/login')
    }
  }

  const apiFetch = <T>(url: string, options?: FetchOptions): Promise<T> =>
    fetchWithRefresh<T>(
      (u, o) => $fetch(u, o as never),
      refresh,
      onAuthFailure,
      url,
      options
    )

  return { apiFetch }
}
