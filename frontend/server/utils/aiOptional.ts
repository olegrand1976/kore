type UpstreamFetchError = {
  statusCode?: number
  data?: {
    error?: {
      code?: string
      message?: string
    }
  }
}

/** True when the upstream AI endpoint is unavailable because AI is disabled (optional feature). */
export function isAiOptionalUnavailableError(err: unknown): boolean {
  const e = err as UpstreamFetchError
  if (e.statusCode !== 403) return false
  const apiError = e.data?.error
  if (apiError?.code !== 'FORBIDDEN') return false
  const message = (apiError.message ?? '').toLowerCase()
  return message.includes('ai assistance disabled') || message.includes('ai capability disabled')
}
