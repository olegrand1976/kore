import { isAbsenceSourceType } from '~/utils/craAbsence'
import { formatHoursValue } from '~/utils/craDuration'

export type DayRowLike = {
  key?: string
  sourceType: string
  hours: string
  origin?: string
  comment?: string
  billable?: boolean
  workRefType?: string
  workRefId?: string
}

export function withManualOrigin<T extends { origin: string }>(row: T): T {
  return row.origin === 'manual' ? row : { ...row, origin: 'manual' }
}

export function unlockHolidayPrefillRows<T extends DayRowLike & { origin: string }>(rows: T[]): T[] {
  return rows.map((row) =>
    row.sourceType === 'holiday' && row.origin === 'prefill' ? withManualOrigin(row) : row
  )
}

/** Journée entièrement en absence (bandeau « jour non presté »). */
export function isFullAbsenceDay(
  rows: DayRowLike[],
  hoursToMinutes: (hours: string) => number
): boolean {
  const absenceRows = rows.filter((row) => isAbsenceSourceType(row.sourceType))
  if (absenceRows.length === 0) return false
  if (rows.some((row) => !isAbsenceSourceType(row.sourceType) && hoursToMinutes(row.hours) > 0)) {
    return false
  }
  if (rows.some((row) => !isAbsenceSourceType(row.sourceType))) return false
  if (absenceRows.some((row) => hoursToMinutes(row.hours) > 0)) return false
  return true
}

export function partialAbsenceHoursLabel(capacityMinutes: number): string {
  return formatHoursValue(capacityMinutes / 2 / 60)
}

/**
 * Clés des lignes modifiées par rapport à l'état enregistré, jour par jour.
 *
 * `saved` est dérivé de la dernière réponse serveur, `draft` porte les éditions en
 * cours : la comparaison par empreinte évite de maintenir une baseline séparée.
 * Une ligne ajoutée n'est signalée que si elle porte des heures ou un commentaire,
 * sinon la ligne vide insérée d'office sur chaque jour serait toujours « modifiée ».
 */
export function dirtyRowKeys(
  draft: Map<string, DayRowLike[]>,
  saved: Map<string, DayRowLike[]>,
  hoursToMinutes: (hours: string) => number
): Set<string> {
  const dirty = new Set<string>()
  for (const [day, rows] of draft) {
    const savedByKey = new Map((saved.get(day) ?? []).map((row) => [row.key, row]))
    for (const row of rows) {
      if (!row.key) continue
      const base = savedByKey.get(row.key)
      if (!base) {
        if (hoursToMinutes(row.hours) > 0 || (row.comment ?? '').trim() !== '') dirty.add(row.key)
        continue
      }
      if (rowsSnapshot([row]) !== rowsSnapshot([base])) dirty.add(row.key)
    }
  }
  return dirty
}

/**
 * Empreinte d'une journée : doit couvrir tous les champs éditables d'une ligne,
 * sinon une modification (ex. commentaire seul) est considérée comme un no-op
 * et n'est jamais remontée au parent — donc jamais enregistrée.
 */
export function rowsSnapshot(rows: DayRowLike[]): string {
  return JSON.stringify(
    rows.map((row) => [
      row.key ?? row.sourceType,
      row.hours,
      row.origin ?? '',
      row.comment ?? '',
      row.billable ?? true,
      row.workRefType ?? '',
      row.workRefId ?? ''
    ])
  )
}
