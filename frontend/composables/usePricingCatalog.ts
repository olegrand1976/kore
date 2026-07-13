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

const EDITION_ORDER: EditionCode[] = ['starter', 'pro', 'enterprise']

function normalizeModuleCodes(codes: string[]): Set<string> {
  return new Set(codes.map((code) => code.toLowerCase()).filter(Boolean))
}

/** Finds the highest edition whose module set matches or is covered by active modules. */
export function matchEdition(activeCodes: string[], editions: EditionPrice[]): EditionPrice | null {
  const active = normalizeModuleCodes(activeCodes)
  if (active.size === 0 || editions.length === 0) {
    return null
  }

  for (let i = EDITION_ORDER.length - 1; i >= 0; i -= 1) {
    const edition = editions.find((item) => item.code === EDITION_ORDER[i])
    if (!edition) {
      continue
    }
    const editionModules = normalizeModuleCodes(edition.modules)
    const exactMatch =
      active.size === editionModules.size && [...active].every((code) => editionModules.has(code))
    const coveredMatch = [...editionModules].every((code) => active.has(code))
    if (exactMatch || coveredMatch) {
      return edition
    }
  }

  return null
}

/** Suggests the next edition tier when the tenant can extend module access. */
export function suggestUpgradeEdition(
  current: EditionPrice | null,
  editions: EditionPrice[]
): EditionPrice | null {
  if (editions.length === 0) {
    return null
  }
  if (!current) {
    return editions.find((edition) => edition.code === 'pro') ?? editions[0] ?? null
  }
  const idx = EDITION_ORDER.indexOf(current.code)
  if (idx < 0 || idx >= EDITION_ORDER.length - 1) {
    return null
  }
  return editions.find((edition) => edition.code === EDITION_ORDER[idx + 1]) ?? null
}

export function modulesMissingFromEdition(activeCodes: string[], edition: EditionPrice): string[] {
  const active = normalizeModuleCodes(activeCodes)
  return edition.modules.filter((code) => !active.has(code.toLowerCase()))
}
