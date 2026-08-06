export type OrgApplication = {
  id?: string
  ID?: string
  libelle?: string
  Libelle?: string
  proprietaire?: string
  Proprietaire?: string
  modeFacturation?: string
  ModeFacturation?: string
  uoActivee?: boolean
  UOActivee?: boolean
  chefUtilisateurId?: string
  ChefUtilisateurID?: string
  budgetDefautId?: string
  BudgetDefautID?: string
  active?: boolean
  Active?: boolean
  serviceId?: string
  ServiceID?: string
}

export type ApplicationWriteBody = {
  serviceId?: string
  libelle?: string
  proprietaire?: string
  modeFacturation?: string
  uoActivee?: boolean
  chefUtilisateurId?: string | null
  budgetDefautId?: string | null
  active?: boolean
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
  return app?.serviceId ?? app?.ServiceID ?? ''
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

  const create = async (body: ApplicationWriteBody & { serviceId: string; libelle: string }) => {
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
    pickAppMode,
    pickAppChefId,
    pickAppBudgetDefautId,
    filterByApplicationId
  }
}
