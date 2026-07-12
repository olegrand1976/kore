export type ModulePrice = {
  code: string
  name: string
  description: string
  unitAmount: number
}

export type EditionCode = 'starter' | 'pro' | 'enterprise'

export type EditionPrice = {
  code: EditionCode
  name: string
  description: string
  unitAmount: number
  modules: string[]
  highlight: boolean
}

type RawModulePrice = {
  code?: string
  Code?: string
  name?: string
  Name?: string
  description?: string
  Description?: string
  unitAmount?: number
  UnitAmount?: number
}

type RawEditionPrice = {
  code?: string
  Code?: string
  name?: string
  Name?: string
  description?: string
  Description?: string
  unitAmount?: number
  UnitAmount?: number
  modules?: string[]
  Modules?: string[]
  highlight?: boolean
  Highlight?: boolean
}

function catalogFromPayload(payload: unknown) {
  const root = payload as {
    data?: {
      catalog?: {
        modules?: RawModulePrice[]
        Modules?: RawModulePrice[]
        editions?: RawEditionPrice[]
        Editions?: RawEditionPrice[]
      }
      modules?: RawModulePrice[]
      Modules?: RawModulePrice[]
      editions?: RawEditionPrice[]
      Editions?: RawEditionPrice[]
    }
  }
  const inner = root?.data
  const catalog = inner?.catalog ?? inner
  return catalog ?? {}
}

function normalizeModulePrice(m: RawModulePrice): ModulePrice {
  return {
    code: String(m.code ?? m.Code ?? ''),
    name: String(m.name ?? m.Name ?? ''),
    description: String(m.description ?? m.Description ?? ''),
    unitAmount: Number(m.unitAmount ?? m.UnitAmount ?? 0)
  }
}

function normalizeEditionCode(raw: string): EditionCode | null {
  switch (raw) {
    case 'starter':
    case 'pro':
    case 'enterprise':
      return raw
    default:
      return null
  }
}

function normalizeEditionPrice(raw: RawEditionPrice): EditionPrice | null {
  const code = normalizeEditionCode(String(raw.code ?? raw.Code ?? '').toLowerCase())
  if (!code) {
    return null
  }
  const modules = (raw.modules ?? raw.Modules ?? []).map((item) => String(item))
  return {
    code,
    name: String(raw.name ?? raw.Name ?? ''),
    description: String(raw.description ?? raw.Description ?? ''),
    unitAmount: Number(raw.unitAmount ?? raw.UnitAmount ?? 0),
    modules,
    highlight: Boolean(raw.highlight ?? raw.Highlight)
  }
}

/** Extracts module prices from `/api/public/pricing` (envelope `data.catalog.modules`). */
export function parsePricingModules(payload: unknown): ModulePrice[] {
  const catalog = catalogFromPayload(payload)
  const raw =
    catalog.modules ??
    catalog.Modules ??
    []
  return raw.map(normalizeModulePrice).filter((m) => m.code !== '')
}

/** Extracts edition bundles from `/api/public/pricing` (envelope `data.catalog.editions`). */
export function parsePricingEditions(payload: unknown): EditionPrice[] {
  const catalog = catalogFromPayload(payload)
  const raw =
    catalog.editions ??
    catalog.Editions ??
    []
  const editions = raw
    .map(normalizeEditionPrice)
    .filter((edition): edition is EditionPrice => edition !== null)

  const order: EditionCode[] = ['starter', 'pro', 'enterprise']
  return editions.sort((a, b) => order.indexOf(a.code) - order.indexOf(b.code))
}

export function minEditionPrice(payload: unknown): number | null {
  const editions = parsePricingEditions(payload)
  if (editions.length === 0) {
    return null
  }
  return Math.min(...editions.map((edition) => edition.unitAmount))
}
