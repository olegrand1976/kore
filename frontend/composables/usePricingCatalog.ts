export type ModulePrice = {
  code: string
  name: string
  description: string
  unitAmount: number
}

/** Extracts module prices from `/api/public/pricing` (envelope `data.catalog.modules`). */
export function parsePricingModules(payload: unknown): ModulePrice[] {
  const root = payload as {
    data?: {
      catalog?: { modules?: RawModulePrice[]; Modules?: RawModulePrice[] }
      modules?: RawModulePrice[]
      Modules?: RawModulePrice[]
    }
  }
  const inner = root?.data
  const catalog = inner?.catalog ?? inner
  const raw =
    catalog?.modules ??
    catalog?.Modules ??
    inner?.modules ??
    inner?.Modules ??
    []
  return raw.map(normalizeModulePrice).filter((m) => m.code !== '')
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

function normalizeModulePrice(m: RawModulePrice): ModulePrice {
  return {
    code: String(m.code ?? m.Code ?? ''),
    name: String(m.name ?? m.Name ?? ''),
    description: String(m.description ?? m.Description ?? ''),
    unitAmount: Number(m.unitAmount ?? m.UnitAmount ?? 0)
  }
}
