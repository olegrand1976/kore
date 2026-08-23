export type CraAnomalyItem = {
  code?: string
  message?: string
  day?: string
}

/** Normalises Go/BFF payloads into user-facing anomaly messages. */
export function normalizeAnomalyMessages(res: unknown): string[] {
  const payload = (res as { data?: unknown })?.data ?? res
  const list = Array.isArray(payload)
    ? payload
    : (payload as { anomalies?: unknown[] } | null)?.anomalies ?? []
  if (!Array.isArray(list)) return []
  return list.map((item) => {
    if (typeof item === 'string') return item
    const anomaly = item as CraAnomalyItem
    const message = anomaly.message?.trim()
    if (message) return message
    const code = anomaly.code?.trim()
    if (code) return code
    return String(item)
  })
}
