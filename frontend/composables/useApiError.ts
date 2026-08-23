type ApiErrorBody = {
  message?: string
  error?: string | { message?: string; code?: string }
}

function asMessage(value: unknown): string | undefined {
  if (typeof value === 'string' && value.trim()) return value
  if (value && typeof value === 'object' && 'message' in value) {
    const nested = (value as { message?: unknown }).message
    if (typeof nested === 'string' && nested.trim()) return nested
  }
  return undefined
}

export function extractFetchError(err: unknown, fallback = 'Une erreur est survenue'): string {
  if (err && typeof err === 'object') {
    const e = err as {
      data?: ApiErrorBody
      statusMessage?: string
      message?: string
    }
    return (
      asMessage(e.data?.error) ??
      asMessage(e.data?.message) ??
      asMessage(e.statusMessage) ??
      asMessage(e.message) ??
      fallback
    )
  }
  return fallback
}

export function extractFetchErrorCode(err: unknown): string | undefined {
  if (err && typeof err === 'object') {
    const e = err as { data?: ApiErrorBody }
    const error = e.data?.error
    if (error && typeof error === 'object' && typeof error.code === 'string') {
      return error.code
    }
  }
  return undefined
}

export function useApiError() {
  return { extractFetchError }
}
