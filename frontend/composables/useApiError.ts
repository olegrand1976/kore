export function extractFetchError(err: unknown, fallback = 'Une erreur est survenue'): string {
  if (err && typeof err === 'object') {
    const e = err as { data?: { message?: string; error?: string }; statusMessage?: string; message?: string }
    return e.data?.message ?? e.data?.error ?? e.statusMessage ?? e.message ?? fallback
  }
  return fallback
}

export function useApiError() {
  return { extractFetchError }
}
