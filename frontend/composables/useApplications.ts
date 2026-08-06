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
  active?: boolean
  Active?: boolean
  serviceId?: string
  ServiceID?: string
}

export function useApplications() {
  const { apiFetch } = useApiFetch()
  const pickAppId = (app: OrgApplication) => app.id ?? app.ID ?? ''
  const pickAppLabel = (app: OrgApplication | undefined | null) => app?.libelle ?? app?.Libelle ?? ''
  const pickAppClient = (app: OrgApplication | undefined | null) => app?.proprietaire ?? app?.Proprietaire ?? ''
  const pickAppActive = (app: OrgApplication | undefined | null) =>
    app?.active ?? app?.Active ?? true

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

  const update = async (id: string, body: { libelle?: string; active?: boolean }) => {
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
    update,
    deactivate,
    activate,
    appById,
    pickAppId,
    pickAppLabel,
    pickAppClient,
    pickAppActive
  }
}
