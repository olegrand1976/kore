import { formatUserDisplayName } from '~/composables/useUserDisplay'

export type CraUserFilterSource = {
  id: string
  prenom?: string
  nom?: string
  login?: string
  active?: boolean
  /** When false, excluded from the CRA collaborator filter (unless present in extras). */
  craRequis?: boolean
}

export type CraUserFilterOption = { value: string; label: string }

function isActive(value: unknown): boolean {
  return value !== false
}

function isCraRequis(value: unknown): boolean {
  // Missing field → treat as required (DB default TRUE / older payloads).
  return value !== false
}

/** Build collaborator select options from the org user directory (not from filtered CRA rows). */
export function buildCraUserFilterOptions(
  users: CraUserFilterSource[],
  extras: Array<{ id: string; label: string }> = [],
  sessionUserId = '',
  sessionLabel = ''
): CraUserFilterOption[] {
  const byId = new Map<string, string>()

  for (const u of users) {
    const id = String(u.id ?? '').trim()
    if (!id) continue
    if (!isActive(u.active)) continue
    if (!isCraRequis(u.craRequis)) continue
    byId.set(id, formatUserDisplayName(u.prenom, u.nom, u.login) || id)
  }

  for (const extra of extras) {
    const id = String(extra.id ?? '').trim()
    if (!id || byId.has(id)) continue
    byId.set(id, extra.label || id)
  }

  if (sessionUserId && !byId.has(sessionUserId)) {
    byId.set(sessionUserId, sessionLabel || sessionUserId)
  }

  return [...byId.entries()]
    .map(([value, label]) => ({ value, label }))
    .sort((a, b) => a.label.localeCompare(b.label, undefined, { sensitivity: 'base' }))
}
