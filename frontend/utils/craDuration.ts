export function parseLineDurationMinutes(line: Record<string, unknown>): number {
  const raw = line.duration ?? line.Duration
  if (typeof raw === 'number' && Number.isFinite(raw)) {
    return Math.max(0, Math.round(raw))
  }
  if (typeof raw === 'string') {
    const parsed = Number(raw)
    return Number.isFinite(parsed) ? Math.max(0, Math.round(parsed)) : 0
  }
  if (raw && typeof raw === 'object') {
    const obj = raw as Record<string, unknown>
    const minutes = obj.minutes ?? obj.Minutes
    const parsed = Number(minutes)
    return Number.isFinite(parsed) ? Math.max(0, Math.round(parsed)) : 0
  }
  return 0
}

export function safeMinutes(value: unknown): number {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? Math.max(0, parsed) : 0
}
