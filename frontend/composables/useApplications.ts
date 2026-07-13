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
}

export function useApplications() {
  const pickAppId = (app: OrgApplication) => app.id ?? app.ID ?? ''
  const pickAppLabel = (app: OrgApplication | undefined | null) => app?.libelle ?? app?.Libelle ?? ''
  const pickAppClient = (app: OrgApplication | undefined | null) => app?.proprietaire ?? app?.Proprietaire ?? ''

  const list = async () => {
    const res = await $fetch<{ data?: OrgApplication[] }>('/api/org/applications')
    return res?.data ?? []
  }

  const get = async (id: string) => {
    const res = await $fetch<{ data?: OrgApplication }>(`/api/org/applications/${id}`)
    return (res?.data ?? res) as OrgApplication
  }

  const appById = (apps: OrgApplication[]) => {
    const map = new Map<string, OrgApplication>()
    for (const app of apps) {
      const id = pickAppId(app)
      if (id) map.set(id, app)
    }
    return map
  }

  return { list, get, appById, pickAppId, pickAppLabel, pickAppClient }
}
