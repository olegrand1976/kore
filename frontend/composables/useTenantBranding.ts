export function useTenantBranding() {
  const branding = useState('tenant-branding', () => ({
    logoUrl: null as string | null,
    raisonSociale: '',
    societeId: null as string | null
  }))

  const resolveLogoUrl = (logo?: string, tenantId?: string) => {
    if (!logo) return null
    if (logo.startsWith('blob:') || logo.startsWith('/api/org/')) return logo
    const id = tenantId || logo.match(/branding\/logo\/([0-9a-f-]+)/i)?.[1]
    if (id) return `/api/org/branding/logo/${id}`
    if (logo.startsWith('http')) return logo
    return null
  }

  const fetchBranding = async () => {
    try {
      const res = await $fetch<{ data: Array<{ id: string; raisonSociale: string; logo?: string; tenantId?: string }> }>('/api/org/societes')
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
