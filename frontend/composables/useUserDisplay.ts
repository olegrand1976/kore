import { formatLeaveUserLogin } from '~/composables/useLeave'

export function formatUserDisplayName(
  prenom?: string,
  nom?: string,
  login?: string
): string {
  const p = prenom?.trim()
  const n = nom?.trim()
  if (p && n) return `${p} ${n}`
  if (p || n) return p || n || '—'
  return formatLeaveUserLogin(login ?? '')
}
