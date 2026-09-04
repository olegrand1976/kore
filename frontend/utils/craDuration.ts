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

/**
 * Formate un nombre d'heures sur 2 décimales, sans zéros de fin :
 * 6.25 → "6.25", 7.5 → "7.5", 8 → "8".
 *
 * 2 décimales = 0,6 min de précision : l'aller-retour minutes → libellé → minutes
 * est stable pour toute valeur entière de minutes, ce qui évite la dérive de la
 * valeur réinjectée dans les champs de saisie du CRA.
 */
export function formatHoursValue(hours: number): string {
  const n = Number(hours)
  if (!Number.isFinite(n)) return '0'
  return String(Math.round(n * 100) / 100)
}
