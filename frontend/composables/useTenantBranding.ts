const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i

/** Resolve a browser-safe logo URL from API logo path / tenant id. */
export function resolveLogoUrl(logo?: string | null, tenantId?: unknown): string | null {
  if (!logo) return null
  if (logo.startsWith('blob:') || logo.startsWith('/api/org/')) return logo

  const fromPath = logo.match(/branding\/logo\/([0-9a-f-]+)/i)?.[1]
  const fromTenant = typeof tenantId === 'string' && UUID_RE.test(tenantId) ? tenantId : null
  // Prefer path UUID: legacy payloads may send tenantId as `{}` (empty Go struct JSON).
  const id = fromPath || fromTenant
  if (id) return `/api/org/branding/logo/${id}`
  if (logo.startsWith('http')) return logo
  return null
}

export function useTenantBranding() {
  const { apiFetch } = useApiFetch()
  const branding = useState('tenant-branding', () => ({
    logoUrl: null as string | null,
    raisonSociale: '',
    societeId: null as string | null
  }))

  const fetchBranding = async () => {
    try {
      const res = await apiFetch<{ data: Array<{ id: string; raisonSociale: string; logo?: string; tenantId?: unknown }> }>('/api/org/societes')
      const first = res.data?.[0]
      if (first) {
        branding.value = {
          logoUrl: resolveLogoUrl(first.logo, first.tenantId),
          raisonSociale: first.raisonSociale,
          societeId: first.id
        }
      }
    } catch {
      // fallback Kore
    }
  }

  return { branding, fetchBranding }
}
