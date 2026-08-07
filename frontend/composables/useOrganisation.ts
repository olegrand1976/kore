// Hiérarchie organisation : Société → Site → Service → Application → Équipe.
// L'équipe est le pivot qui rattache un collaborateur à son application de travail
// (org.users.equipe_id → org.equipes.application_id).

export type OrgSociete = {
  id?: string
  ID?: string
  raisonSociale?: string
  RaisonSociale?: string
}

export type OrgSite = {
  id?: string
  ID?: string
  societeId?: string
  SocieteID?: string
  libelle?: string
  Libelle?: string
  pays?: string
}

export type OrgService = {
  id?: string
  ID?: string
  siteId?: string
  SiteID?: string
  siteLabel?: string
  societeId?: string
  libelle?: string
  Libelle?: string
  type?: string
  responsableId?: string
}

export type OrgApplication = {
  id?: string
  ID?: string
  serviceId?: string
  ServiceID?: string
  siteIds?: string[]
  SiteIDs?: string[]
  serviceIds?: string[]
  ServiceIDs?: string[]
  equipeIds?: string[]
  EquipeIDs?: string[]
  libelle?: string
  Libelle?: string
  active?: boolean
  Active?: boolean
}

export type OrgEquipe = {
  id?: string
  ID?: string
  applicationId?: string
  ApplicationID?: string
  libelle?: string
  Libelle?: string
  responsableId?: string
}

// Les handlers Go sérialisent tantôt en camelCase (tags json) tantôt avec les noms
// de champs Go exportés selon la structure : on absorbe les deux formes.
export function orgId(item: { id?: string; ID?: string } | undefined): string {
  return item?.id ?? item?.ID ?? ''
}

export function orgLabel(item: { libelle?: string; Libelle?: string } | undefined): string {
  return item?.libelle ?? item?.Libelle ?? ''
}

function unwrap<T>(res: unknown): T[] {
  const payload = (res as { data?: T[] })?.data ?? res
  return Array.isArray(payload) ? payload : []
}

export type EquipeOption = { value: string; label: string }

/**
 * Construit les options d'un sélecteur d'équipe en affichant « Équipe — Application ».
 * L'application lève l'ambiguïté entre équipes homonymes rattachées à des périmètres
 * différents. Fonction pure : testable sans les globales Nuxt.
 */
export function buildEquipeOptions(
  equipes: OrgEquipe[],
  applications: OrgApplication[]
): EquipeOption[] {
  const appLabels = new Map(applications.map((a) => [orgId(a), orgLabel(a)]))
  return equipes.map((equipe) => {
    const appLabel = appLabels.get(equipe.applicationId ?? equipe.ApplicationID ?? '') ?? ''
    const equipeLabel = orgLabel(equipe)
    return {
      value: orgId(equipe),
      label: appLabel ? `${equipeLabel} — ${appLabel}` : equipeLabel
    }
  })
}

export function useOrganisation() {
  const { apiFetch } = useApiFetch()

  const listSocietes = async () => unwrap<OrgSociete>(await apiFetch('/api/org/societes'))
  const listSites = async () => unwrap<OrgSite>(await apiFetch('/api/org/sites'))
  const listServices = async () => unwrap<OrgService>(await apiFetch('/api/org/services'))
  const listApplications = async () =>
    unwrap<OrgApplication>(await apiFetch('/api/org/applications?active=all'))
  const listEquipes = async (opts?: { applicationId?: string }) => {
    const qs = opts?.applicationId ? `?applicationId=${encodeURIComponent(opts.applicationId)}` : ''
    return unwrap<OrgEquipe>(await apiFetch(`/api/org/equipes${qs}`))
  }

  const createSite = (body: { societeId: string; libelle: string; pays?: string }) =>
    apiFetch('/api/org/sites', { method: 'POST', body })

  const updateSite = (id: string, body: { libelle: string }) =>
    apiFetch(`/api/org/sites/${id}`, { method: 'PUT', body })

  const createService = (body: {
    siteId: string
    libelle: string
    type?: string
    responsableId: string
  }) => apiFetch('/api/org/services', { method: 'POST', body })

  const updateService = (id: string, body: { libelle: string }) =>
    apiFetch(`/api/org/services/${id}`, { method: 'PUT', body })

  const {
    create: createApplication,
    update: updateApplication,
    deactivate: deactivateApplication,
    activate: activateApplication
  } = useApplications()

  const createEquipe = (body: { applicationId: string; libelle: string; responsableId?: string }) =>
    apiFetch('/api/org/equipes', { method: 'POST', body })

  const listClients = async () =>
    unwrap<{
      id?: string
      ID?: string
      raisonSociale?: string
      RaisonSociale?: string
      tva?: string
      pays?: string
      adresse?: string
      adresseNumero?: string
      adresseBoite?: string
      codePostal?: string
      ville?: string
      siret?: string
    }>(await apiFetch('/api/org/clients'))

  const createClient = (body: {
    raisonSociale: string
    tva?: string
    pays?: string
    adresse?: string
    adresseNumero?: string
    adresseBoite?: string
    codePostal?: string
    ville?: string
    siret?: string
  }) => apiFetch('/api/org/clients', { method: 'POST', body })

  const updateClient = (
    id: string,
    body: {
      raisonSociale: string
      tva: string
      pays: string
      adresse: string
      adresseNumero: string
      adresseBoite: string
      codePostal: string
      ville: string
      siret: string
    }
  ) => apiFetch(`/api/org/clients/${id}`, { method: 'PUT', body })

  return {
    listSocietes,
    listSites,
    listServices,
    listApplications,
    listEquipes,
    createSite,
    updateSite,
    createService,
    updateService,
    createApplication,
    updateApplication,
    deactivateApplication,
    activateApplication,
    createEquipe,
    listClients,
    createClient,
    updateClient,
    orgId,
    orgLabel,
    buildEquipeOptions
  }
}
