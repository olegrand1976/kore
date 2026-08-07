import { normalizeCountryCode, type CountryCode } from '~/composables/useCountryTimezone'

/**
 * Loads the tenant société country once (first société) for timezone display.
 * Multi-société tenants: first row only — same heuristic as congés / branding.
 */
export function useOrgPays() {
  const { listSocietes } = useOrganisation()

  const { data, pending, refresh } = useAsyncData(
    'org-societe-pays',
    async (): Promise<CountryCode> => {
      const societes = await listSocietes()
      const first = societes[0]
      return normalizeCountryCode(first?.pays ?? first?.Pays)
    },
    { default: () => 'FR' as CountryCode }
  )

  const orgPays = computed(() => data.value ?? ('FR' as CountryCode))

  return { orgPays, pending, refresh }
}
