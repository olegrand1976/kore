/** Client-side TTL cache for CRA gate mode (avoids refetch on every /conges|/tma nav). */
export const CRA_GATE_TTL_MS = 60_000

export type CraGateModeCache = {
  mode: string
  fetchedAt: number
}

export function readCachedCraGateMode(
  cache: CraGateModeCache | null | undefined,
  now = Date.now()
): string | null {
  if (!cache) return null
  if (now - cache.fetchedAt >= CRA_GATE_TTL_MS) return null
  return cache.mode
}

export function writeCraGateModeCache(mode: string, now = Date.now()): CraGateModeCache {
  return { mode, fetchedAt: now }
}
