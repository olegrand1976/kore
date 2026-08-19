export type MissionCommercialPatch = {
  client?: string
  clientId?: string
  technologies: string[]
  responsableClient: string
}

export type PrestationInfoFields = {
  client: string
  mission: string
  clientId: string
  missionId: string
  description: string
  technologies: string[]
  lieu: string
  responsableClient: string
}

export const isManualCommercialEntry = (missionId: string | undefined | null): boolean =>
  !String(missionId ?? '').trim()

export function isKnownMissionLink(
  missionId: string | undefined | null,
  missions: Array<{ id: string }>
): boolean {
  const id = String(missionId ?? '').trim()
  if (!id) return false
  return missions.some((item) => item.id === id)
}

export function prestationInfoComplete(client: string, mission: string): boolean {
  return Boolean(client.trim() && mission.trim())
}

export function unwrapMissionPayload(res: unknown): Record<string, unknown> {
  if (!res || typeof res !== 'object') return {}
  const body = res as { data?: unknown }
  if (body.data && typeof body.data === 'object' && !Array.isArray(body.data)) {
    return body.data as Record<string, unknown>
  }
  return res as Record<string, unknown>
}

export function missionCommercialPatch(raw: Record<string, unknown>): MissionCommercialPatch {
  const clientName = String(raw.clientName ?? raw.ClientName ?? '').trim()
  const clientId = String(raw.clientId ?? raw.ClientID ?? '').trim()
  const techs = raw.technologies ?? raw.Technologies
  const patch: MissionCommercialPatch = {
    technologies: Array.isArray(techs)
      ? techs.map((item) => String(item).trim()).filter(Boolean)
      : [],
    responsableClient: String(raw.clientContact ?? raw.ClientContact ?? '').trim()
  }
  if (clientName) patch.client = clientName
  if (clientId) patch.clientId = clientId
  return patch
}
