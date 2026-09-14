export type MissionPrestationPatch = {
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

export const isManualPrestationEntry = (missionId: string | undefined | null): boolean =>
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

/** Mission link UI: full edits while timesheet editable; fill-once after final if incomplete. */
export function canLinkCraMission(opts: {
  canWrite: boolean
  canEditTimesheet: boolean
  prestationComplete: boolean
}): boolean {
  return opts.canWrite && (opts.canEditTimesheet || !opts.prestationComplete)
}

export function unwrapMissionPayload(res: unknown): Record<string, unknown> {
  if (!res || typeof res !== 'object') return {}
  const body = res as { data?: unknown }
  if (body.data && typeof body.data === 'object' && !Array.isArray(body.data)) {
    return body.data as Record<string, unknown>
  }
  return res as Record<string, unknown>
}

/** Display label for mission option lists: "Title — Client". */
export type MissionOptionFields = {
  id?: string
  title?: string
  /** @deprecated Prefer title; kept for callers that still pass label. */
  label?: string
  clientName?: string
}

export function formatMissionOptionLabel(mission: MissionOptionFields): string {
  const title = (mission.title ?? mission.label ?? '').trim()
  const client = (mission.clientName ?? '').trim()
  if (title && client) return `${title} — ${client}`
  if (title) return title
  if (client) return client
  const id = String(mission.id ?? '').trim()
  return id ? id.slice(0, 8) : ''
}

export function missionTitleFromPayload(raw: Record<string, unknown>): string {
  return String(raw.title ?? raw.Title ?? raw.label ?? raw.Label ?? '').trim()
}

export function mapMissionListItem(item: Record<string, unknown>): {
  id: string
  clientName: string
  clientId: string
  title: string
  label: string
} {
  const id = String(item.id ?? item.ID ?? '').trim()
  const clientName = String(item.clientName ?? item.ClientName ?? '').trim()
  const clientId = String(item.clientId ?? item.ClientID ?? '').trim()
  const title = missionTitleFromPayload(item)
  return {
    id,
    clientName,
    clientId,
    title,
    label: title
  }
}

export function missionPrestationPatch(raw: Record<string, unknown>): MissionPrestationPatch {
  const clientName = String(raw.clientName ?? raw.ClientName ?? '').trim()
  const clientId = String(raw.clientId ?? raw.ClientID ?? '').trim()
  const techs = raw.technologies ?? raw.Technologies
  const patch: MissionPrestationPatch = {
    technologies: Array.isArray(techs)
      ? techs.map((item) => String(item).trim()).filter(Boolean)
      : [],
    responsableClient: String(raw.clientContact ?? raw.ClientContact ?? '').trim()
  }
  if (clientName) patch.client = clientName
  if (clientId) patch.clientId = clientId
  return patch
}
