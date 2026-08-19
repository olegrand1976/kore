import type { MethodologyProfile } from '~/composables/useMethodologyTerms'

export type OrgApplication = {
  id?: string
  ID?: string
  libelle?: string
  Libelle?: string
  proprietaire?: string
  Proprietaire?: string
  modeFacturation?: string
  ModeFacturation?: string
  defaultTjmCents?: number
  DefaultTJMCents?: number
  uoActivee?: boolean
  UOActivee?: boolean
  chefUtilisateurId?: string
  ChefUtilisateurID?: string
  budgetDefautId?: string
  BudgetDefautID?: string
  active?: boolean
  Active?: boolean
  /** @deprecated use serviceIds */
  serviceId?: string
  ServiceID?: string
  siteIds?: string[]
  SiteIDs?: string[]
  serviceIds?: string[]
  ServiceIDs?: string[]
  equipeIds?: string[]
  EquipeIDs?: string[]
  methodologyProfile?: string
  MethodologyProfile?: string
}

export type ApplicationWriteBody = {
  serviceId?: string
  siteIds?: string[]
  serviceIds?: string[]
  equipeIds?: string[]
  libelle?: string
  proprietaire?: string
  modeFacturation?: string
  defaultTjmCents?: number
  uoActivee?: boolean
  chefUtilisateurId?: string | null
  budgetDefautId?: string | null
  active?: boolean
  methodologyProfile?: string
}

export const MODE_FACTURATION_VALUES = ['non', 'forfait', 'temps_passe'] as const
export type ModeFacturation = (typeof MODE_FACTURATION_VALUES)[number]

export function pickAppId(app: OrgApplication | undefined | null) {
  return app?.id ?? app?.ID ?? ''
}

export function pickAppLabel(app: OrgApplication | undefined | null) {
  return app?.libelle ?? app?.Libelle ?? ''
}

export function pickAppClient(app: OrgApplication | undefined | null) {
  return app?.proprietaire ?? app?.Proprietaire ?? ''
}

export function pickAppActive(app: OrgApplication | undefined | null) {
  return app?.active ?? app?.Active ?? true
}

export function pickAppServiceId(app: OrgApplication | undefined | null) {
  const ids = pickAppServiceIds(app)
  return ids[0] ?? ''
}

export function pickUUIDList(raw: string[] | undefined | null): string[] {
  if (!Array.isArray(raw)) return []
  return raw.map(String).filter(Boolean)
}

export function pickAppSiteIds(app: OrgApplication | undefined | null) {
  return pickUUIDList(app?.siteIds ?? app?.SiteIDs)
}

export function pickAppServiceIds(app: OrgApplication | undefined | null) {
  const ids = pickUUIDList(app?.serviceIds ?? app?.ServiceIDs)
  if (ids.length) return ids
  const legacy = app?.serviceId ?? app?.ServiceID ?? ''
  return legacy ? [String(legacy)] : []
}

export function pickAppEquipeIds(app: OrgApplication | undefined | null) {
  return pickUUIDList(app?.equipeIds ?? app?.EquipeIDs)
}

export function summarizeAppShares(app: OrgApplication | undefined | null): {
  sites: number
  services: number
  equipes: number
} {
  return {
    sites: pickAppSiteIds(app).length,
    services: pickAppServiceIds(app).length,
    equipes: pickAppEquipeIds(app).length
  }
}

export function pickAppMode(app: OrgApplication | undefined | null) {
  return app?.modeFacturation ?? app?.ModeFacturation ?? 'temps_passe'
}

export function pickAppChefId(app: OrgApplication | undefined | null) {
  return app?.chefUtilisateurId ?? app?.ChefUtilisateurID ?? ''
}

export function pickAppBudgetDefautId(app: OrgApplication | undefined | null) {
  return app?.budgetDefautId ?? app?.BudgetDefautID ?? ''
}

export function pickAppMethodologyProfile(app: OrgApplication | undefined | null): MethodologyProfile {
  const raw = app?.methodologyProfile ?? app?.MethodologyProfile ?? 'psa'
  if (raw === 'agile_scrum' || raw === 'agile_kanban') return raw
  return 'psa'
}

/** RG-BUD-01: default budget type is stored as "defaut" (legacy "default" accepted in UI). */
export function isDefaultBudgetType(type: string | undefined | null): boolean {
  const normalized = String(type ?? '').toLowerCase()
  return normalized === 'defaut' || normalized === 'default'
}

export function filterByApplicationId<T extends { applicationId?: string; ApplicationID?: string }>(
  items: T[],
  applicationId: string
) {
  return items.filter((item) => (item.applicationId ?? item.ApplicationID ?? '') === applicationId)
}

export function defaultBudgetsForApplication<
  T extends {
    applicationId?: string
    ApplicationID?: string
    type?: string
    Type?: string
  }
>(items: T[], applicationId: string): T[] {
  return filterByApplicationId(items, applicationId).filter((item) =>
    isDefaultBudgetType(item.type ?? item.Type)
  )
}

/** Keep selection only if it is a known default-type budget; otherwise clear (stale FK). */
export function coerceBudgetDefautId(selected: string, allowedIds: readonly string[]): string {
  if (!selected) return ''
  return allowedIds.includes(selected) ? selected : ''
}

export function useApplications() {
  const { apiFetch } = useApiFetch()

  const list = async (opts?: { active?: 'true' | 'false' | 'all' }) => {
    const active = opts?.active ?? 'true'
    const qs = active === 'true' ? '' : `?active=${active}`
    const res = await apiFetch<{ data?: OrgApplication[] }>(`/api/org/applications${qs}`)
    return res?.data ?? []
  }

  const get = async (id: string) => {
    const res = await apiFetch<{ data?: OrgApplication }>(`/api/org/applications/${id}`)
    return (res?.data ?? res) as OrgApplication
  }

  const create = async (
    body: ApplicationWriteBody & { libelle: string } & (
      | { serviceId: string }
      | { serviceIds: string[] }
      | { siteIds: string[] }
      | { equipeIds: string[] }
    )
  ) => {
    return apiFetch<{ data?: OrgApplication }>('/api/org/applications', {
      method: 'POST',
      body
    })
  }

  const update = async (id: string, body: ApplicationWriteBody) => {
    return apiFetch<{ data?: OrgApplication }>(`/api/org/applications/${id}`, {
      method: 'PUT',
      body
    })
  }

  const deactivate = async (id: string) => {
    return apiFetch(`/api/org/applications/${id}/deactivate`, { method: 'PATCH' })
  }

  const activate = async (id: string) => {
    return apiFetch(`/api/org/applications/${id}/activate`, { method: 'PATCH' })
  }

  const appById = (apps: OrgApplication[]) => {
    const map = new Map<string, OrgApplication>()
    for (const app of apps) {
      const id = pickAppId(app)
      if (id) map.set(id, app)
    }
    return map
  }

  return {
    list,
    get,
    create,
    update,
    deactivate,
    activate,
    appById,
    pickAppId,
    pickAppLabel,
    pickAppClient,
    pickAppActive,
    pickAppServiceId,
    pickAppSiteIds,
    pickAppServiceIds,
    pickAppEquipeIds,
    summarizeAppShares,
    pickAppMode,
    pickAppChefId,
    pickAppBudgetDefautId,
    pickAppMethodologyProfile,
    filterByApplicationId
  }
}
