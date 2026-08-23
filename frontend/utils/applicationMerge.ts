import type { MethodologyProfile } from '~/composables/useMethodologyTerms'

export type MergeApplicationRow = {
  id: string
  libelle: string
  methodologyProfile: MethodologyProfile
  taigaLinked: boolean
}

export function countTaigaLinked(rows: MergeApplicationRow[]): number {
  return rows.filter((row) => row.taigaLinked).length
}

export function canMergeApplications(rows: MergeApplicationRow[]): boolean {
  return rows.length === 2 && countTaigaLinked(rows) < 2
}

export function isMergeReferenceLocked(rows: MergeApplicationRow[]): boolean {
  return countTaigaLinked(rows) === 1
}

export function defaultMergeReferenceId(rows: MergeApplicationRow[]): string {
  const taiga = rows.find((row) => row.taigaLinked)
  if (taiga) return taiga.id
  return rows[0]?.id ?? ''
}

export function mergeAbsorbedId(rows: MergeApplicationRow[], referenceId: string): string {
  return rows.find((row) => row.id !== referenceId)?.id ?? ''
}

export function hasMergeMethodologyMismatch(rows: MergeApplicationRow[], referenceId: string): boolean {
  const reference = rows.find((row) => row.id === referenceId)
  if (!reference) return false
  return rows.some((row) => row.methodologyProfile !== reference.methodologyProfile)
}
