/** Routes où la création d'activité est bloquée si le CRA du mois est incomplet (mode block). */
export function isCraGateBlockedPath(path: string): boolean {
  const normalized = path.replace(/\/$/, '') || '/'
  if (normalized === '/conges') return true
  if (normalized === '/tma') return true
  if (/^\/tma\/[^/]+\/new(\/|$)/.test(normalized)) return true
  if (normalized.startsWith('/conges/demande')) return true
  return false
}
