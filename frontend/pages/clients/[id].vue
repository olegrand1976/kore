<template>
  <div>
    <AppPageHeader :title="title" :subtitle="$t('fiche.client_title')">
      <template #actions>
        <AppButton
          v-if="canEdit && client && !editing"
          variant="secondary"
          size="sm"
          type="button"
          class="client-edit-btn"
          @click="startEdit"
        >
          <AppIcon name="edit" /> {{ $t('clients.edit_billing') }}
        </AppButton>
        <AppButton variant="ghost" size="sm" to="/clients">
          <AppIcon name="arrow_back" /> {{ $t('clients.back_list') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('fiche.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="error" padding="lg">
      <AppEmptyState icon="error" :title="$t('fiche.not_found')" />
    </AppCard>

    <template v-else-if="client">
      <AppKpiGrid compact>
        <AppKpiCard
          icon="corporate_fare"
          tone="gold"
          :value="title"
          :label="$t('fiche.col_company')"
        />
        <AppKpiCard
          icon="receipt_long"
          tone="blue"
          :value="client.tva || $t('fiche.none')"
          :label="$t('fiche.col_vat')"
        />
        <AppKpiCard
          icon="groups"
          tone="success"
          :value="contacts.length"
          :label="$t('fiche.section_contacts')"
        />
        <AppKpiCard
          icon="work"
          tone="warn"
          :value="missions.length"
          :label="$t('fiche.section_missions')"
        />
      </AppKpiGrid>

      <AppCard padding="lg" class="fiche-block">
        <div class="fiche-section-head">
          <h3 class="fiche-section-title">{{ $t('clients.section_billing') }}</h3>
        </div>

        <form v-if="editing" class="client-edit-form" @submit.prevent="submitEdit">
          <ClientBillingForm v-model="editForm" id-prefix="client-edit" />
          <p v-if="editError" class="flash flash--error" role="alert">{{ editError }}</p>
          <div class="client-edit-form__actions">
            <AppButton variant="ghost" type="button" :disabled="saving" @click="cancelEdit">
              {{ $t('common.cancel') }}
            </AppButton>
            <AppButton variant="primary" type="submit" :loading="saving">
              {{ $t('common.save') }}
            </AppButton>
          </div>
        </form>

        <dl v-else class="fiche-dl fiche-dl--grid">
          <div>
            <dt>{{ $t('fiche.col_company') }}</dt>
            <dd>{{ title }}</dd>
          </div>
          <div>
            <dt>{{ $t('fiche.col_vat') }}</dt>
            <dd>{{ client.tva || $t('fiche.none') }}</dd>
          </div>
          <div>
            <dt>{{ registryLabel }}</dt>
            <dd>{{ client.siret || $t('fiche.none') }}</dd>
          </div>
          <div class="fiche-dl__wide">
            <dt>{{ $t('org.address') }}</dt>
            <dd>{{ formattedAddress }}</dd>
          </div>
          <div>
            <dt>{{ $t('fiche.col_created') }}</dt>
            <dd>{{ formatDate(client.createdAt) }}</dd>
          </div>
        </dl>
      </AppCard>

      <AppCard padding="none" class="fiche-table-wrap fiche-block">
        <div class="fiche-table-head">
          <h3 class="fiche-section-title">{{ $t('fiche.section_contacts') }}</h3>
        </div>
        <div class="fiche-table-toolbar">
          <AppListToolbar
            :filters="contactListFilters"
            :filter-values="contactFilterValues"
            :sort-keys="contactSortKeys"
            :sort-key="contactSortKey"
            :sort-dir="contactSortDir"
            :has-active-filters="contactHasActiveFilters"
            @update:filter="setContactFilter"
            @update:sort-key="setContactSort($event)"
            @update:sort-dir="setContactSortDir"
            @reset="resetContactFilters"
          />
        </div>
        <AppTable
          :columns="contactColumns"
          :rows="contactDisplayRows"
          row-key="email"
          :empty-title="contactHasActiveFilters ? $t('common.list.no_results') : $t('fiche.contacts_empty')"
        >
          <template #cell-name="{ row }">
            <span class="fiche-strong">{{ row.name }}</span>
          </template>
          <template #cell-email="{ value }">
            <a v-if="value" :href="`mailto:${value}`" class="fiche-link">{{ value }}</a>
            <span v-else>{{ $t('fiche.none') }}</span>
          </template>
        </AppTable>
      </AppCard>

      <AppCard padding="none" class="fiche-table-wrap">
        <div class="fiche-table-head">
          <h3 class="fiche-section-title">{{ $t('fiche.section_missions') }}</h3>
        </div>
        <div class="fiche-table-toolbar">
          <AppListToolbar
            :filters="missionListFilters"
            :filter-values="missionFilterValues"
            :sort-keys="missionSortKeys"
            :sort-key="missionSortKey"
            :sort-dir="missionSortDir"
            :has-active-filters="missionHasActiveFilters"
            @update:filter="setMissionFilter"
            @update:sort-key="setMissionSort($event)"
            @update:sort-dir="setMissionSortDir"
            @reset="resetMissionFilters"
          />
        </div>
        <AppTable
          :columns="missionColumns"
          :rows="missionDisplayRows"
          row-key="id"
          :loading="missionsPending"
          :empty-title="missionHasActiveFilters ? $t('common.list.no_results') : $t('fiche.missions_empty')"
        >
          <template #cell-status="{ value }">
            <AppBadge :variant="missionStatusVariant(String(value))">
              {{ missionStatusLabel(String(value)) }}
            </AppBadge>
          </template>
          <template #cell-period="{ row }">
            <span class="fiche-nowrap">{{ row.period }}</span>
          </template>
          <template #cell-tjm="{ value }">
            <span class="fiche-nowrap">{{ value }}</span>
          </template>
          <template #cell-actions="{ row }">
            <AppButton variant="ghost" size="sm" @click="navigateTo(`/missions/${row.id}`)">
              {{ $t('fiche.open_mission') }}
            </AppButton>
          </template>
        </AppTable>
      </AppCard>
    </template>
  </div>
</template>

<script setup lang="ts">
import { applyTextSearch, useListControls } from '~/composables/useListControls'
import {
  clientBillingPayload,
  emptyClientBillingFields,
  normalizeBillingCountry,
  useBillingCountryLabels,
  type ClientBillingFields
} from '~/composables/useBillingCountry'

definePageMeta({ layout: 'default' })

type ClientContact = {
  nom?: string
  prenom?: string
  email?: string
  role?: string
  telephone?: string
}

type ClientDetail = {
  id?: string
  raisonSociale?: string
  tva?: string
  pays?: string
  adresse?: string
  adresseNumero?: string
  adresseBoite?: string
  codePostal?: string
  ville?: string
  siret?: string
  contacts?: ClientContact[]
  createdAt?: string
}

type MissionSummary = {
  id: string
  clientId?: string
  status?: string
  startDate?: string
  endDate?: string
  tjmAmount?: number
  currency?: string
}

const route = useRoute()
const { t } = useI18n()
const { can } = usePermissions()
const { updateClient } = useOrganisation()
const { extractFetchError } = useApiError()
const { formatDate, formatMoney, missionStatusLabel, missionStatusVariant } = useFicheFormat()

const canEdit = computed(() => can('org', 'E'))
const id = computed(() => String(route.params.id ?? ''))

const { data, pending, error, refresh } = await useFetch<ClientDetail>(() => `/api/org/clients/${id.value}`, {
  watch: [id]
})

const { data: missionsData, pending: missionsPending } = await useFetch('/api/ssii/missions', {
  lazy: true,
  server: false
})

const client = computed(() => {
  const payload = (data.value as { data?: ClientDetail })?.data ?? data.value
  return payload && typeof payload === 'object' ? payload : null
})

const title = computed(() => client.value?.raisonSociale ?? '—')

const countryLabel = computed(() => {
  const code = normalizeBillingCountry(client.value?.pays)
  switch (code) {
    case 'FR':
      return t('org.country_fr')
    case 'BE':
      return t('org.country_be')
    case 'MG':
      return t('org.country_mg')
    case 'MA':
      return t('org.country_ma')
    case 'TN':
      return t('org.country_tn')
    case 'CA':
      return t('org.country_ca')
    case '':
      return t('fiche.none')
    default: {
      const _exhaustive: never = code
      return _exhaustive
    }
  }
})

const formattedAddress = computed(() => {
  const c = client.value
  if (!c) return t('fiche.none')
  const street = [c.adresse, c.adresseNumero].filter(Boolean).join(' ')
  const withBox = c.adresseBoite ? (street ? `${street} / ${c.adresseBoite}` : c.adresseBoite) : street
  const city = [c.codePostal, c.ville].filter(Boolean).join(' ')
  const country = countryLabel.value === t('fiche.none') ? '' : countryLabel.value
  const parts = [withBox, city, country].filter(Boolean)
  return parts.length ? parts.join(', ') : t('fiche.none')
})

const paysRef = computed(() => normalizeBillingCountry(client.value?.pays))
const { registryLabel } = useBillingCountryLabels(paysRef)

const editing = ref(false)
const saving = ref(false)
const editError = ref('')
const editForm = reactive<ClientBillingFields>(emptyClientBillingFields())

const fillEditForm = () => {
  const c = client.value
  Object.assign(editForm, emptyClientBillingFields(), {
    raisonSociale: c?.raisonSociale ?? '',
    tva: c?.tva ?? '',
    pays: normalizeBillingCountry(c?.pays),
    adresse: c?.adresse ?? '',
    adresseNumero: c?.adresseNumero ?? '',
    adresseBoite: c?.adresseBoite ?? '',
    codePostal: c?.codePostal ?? '',
    ville: c?.ville ?? '',
    siret: c?.siret ?? ''
  })
}

const startEdit = () => {
  fillEditForm()
  editError.value = ''
  editing.value = true
}

const cancelEdit = () => {
  editing.value = false
  editError.value = ''
}

const submitEdit = async () => {
  saving.value = true
  editError.value = ''
  const payload = clientBillingPayload(editForm)
  if (!payload.raisonSociale) {
    editError.value = t('clients.update_error')
    saving.value = false
    return
  }
  try {
    await updateClient(id.value, payload)
    editing.value = false
    await refresh()
  } catch (err) {
    editError.value = extractFetchError(err, t('clients.update_error'))
  } finally {
    saving.value = false
  }
}

const contacts = computed(() => client.value?.contacts ?? [])

const contactColumns = computed(() => [
  { key: 'name', label: t('fiche.col_contact_name') },
  { key: 'role', label: t('fiche.col_contact_role') },
  { key: 'email', label: t('fiche.col_email') },
  { key: 'phone', label: t('fiche.col_contact_phone') }
])

const contactRows = computed(() =>
  contacts.value.map((c) => ({
    email: c.email ?? `${c.prenom}-${c.nom}`,
    name: [c.prenom, c.nom].filter(Boolean).join(' ') || '—',
    role: c.role || '—',
    phone: c.telephone || '—'
  }))
)

const contactListFilters = computed(() => ({
  q: {
    type: 'search' as const,
    label: t('common.list.search'),
    placeholder: t('fiche.col_contact_name'),
    match: (row: { name: string; role: string; email: string }, query: string) =>
      applyTextSearch(query, row.name, row.role, row.email)
  }
}))

const contactSortKeys = computed(() => [
  { key: 'name', label: t('fiche.col_contact_name'), type: 'string' as const, accessor: (row: { name: string }) => row.name }
])

const {
  filterValues: contactFilterValues,
  sortKey: contactSortKey,
  sortDir: contactSortDir,
  sortedItems: contactSortedItems,
  hasActiveFilters: contactHasActiveFilters,
  setFilter: setContactFilter,
  setSort: setContactSort,
  setSortDir: setContactSortDir,
  resetFilters: resetContactFilters
} = useListControls(contactRows, {
  storageKey: 'client-contacts',
  defaultSort: { key: 'name', dir: 'asc' },
  filters: contactListFilters,
  sortKeys: contactSortKeys
})

const contactDisplayRows = computed(() => contactSortedItems.value)

const missions = computed((): MissionSummary[] => {
  const payload = (missionsData.value as { data?: MissionSummary[] })?.data ?? missionsData.value
  if (!Array.isArray(payload)) return []
  return payload.filter((m) => String(m.clientId ?? '') === id.value)
})

const missionColumns = computed(() => [
  { key: 'period', label: t('fiche.col_period'), nowrap: true },
  { key: 'status', label: t('fiche.col_status') },
  { key: 'tjm', label: t('fiche.col_tjm'), nowrap: true },
  { key: 'actions', label: '' }
])

const formatPeriod = (start?: string, end?: string) => {
  if (!start) return '—'
  const startLabel = formatDate(start)
  if (!end) return startLabel
  return `${startLabel} → ${formatDate(end)}`
}

const missionRows = computed(() =>
  missions.value.map((m) => ({
    id: m.id,
    period: formatPeriod(m.startDate, m.endDate),
    status: m.status ?? '',
    tjm: formatMoney(Number(m.tjmAmount ?? 0), m.currency ?? 'EUR'),
    actions: '',
    searchText: `${formatPeriod(m.startDate, m.endDate)} ${m.status ?? ''}`
  }))
)

const missionListFilters = computed(() => ({
  q: {
    type: 'search' as const,
    label: t('common.list.search'),
    placeholder: t('fiche.section_missions'),
    match: (row: { searchText: string }, query: string) => applyTextSearch(query, row.searchText)
  }
}))

const missionSortKeys = computed(() => [
  { key: 'period', label: t('fiche.col_period'), type: 'string' as const, accessor: (row: { period: string }) => row.period }
])

const {
  filterValues: missionFilterValues,
  sortKey: missionSortKey,
  sortDir: missionSortDir,
  sortedItems: missionSortedItems,
  hasActiveFilters: missionHasActiveFilters,
  setFilter: setMissionFilter,
  setSort: setMissionSort,
  setSortDir: setMissionSortDir,
  resetFilters: resetMissionFilters
} = useListControls(missionRows, {
  storageKey: 'client-missions',
  defaultSort: { key: 'period', dir: 'asc' },
  filters: missionListFilters,
  sortKeys: missionSortKeys
})

const missionDisplayRows = computed(() => missionSortedItems.value)
</script>

<style scoped>
.muted { color: var(--kore-text-muted); }

.flash--error { color: var(--kore-error); }

.fiche-block { margin-bottom: var(--kore-space-lg); }

.fiche-section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--kore-space-md);
  margin-bottom: var(--kore-space-md);
}

.fiche-section-title {
  margin: 0;
  font-size: var(--kore-text-h3);
}

.fiche-dl {
  display: grid;
  gap: var(--kore-space-md);
  margin: 0;
}

.fiche-dl--grid {
  grid-template-columns: 1fr;
}

@media (min-width: 768px) {
  .fiche-dl--grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .fiche-dl__wide {
    grid-column: 1 / -1;
  }
}

.fiche-dl div {
  display: grid;
  gap: var(--kore-space-xs);
}

.fiche-dl dt {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
  font-weight: 500;
}

.fiche-dl dd {
  margin: 0;
  color: var(--kore-text);
  font-weight: 500;
}

.client-edit-form {
  display: grid;
  gap: var(--kore-space-md);
  max-width: var(--kore-form-wide-max);
}

.client-edit-form__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  justify-content: flex-end;
}

.fiche-table-wrap { overflow: hidden; }

.fiche-table-head {
  padding: var(--kore-space-lg) var(--kore-space-lg) 0;
}

.fiche-table-toolbar :deep(.list-toolbar) {
  margin-bottom: 0;
  border: none;
  box-shadow: none;
  padding-top: 0;
}

.fiche-strong { font-weight: 600; }

.fiche-link {
  color: var(--kore-brand-blue);
  text-decoration: none;
}

.fiche-link:hover { text-decoration: underline; }

.fiche-nowrap { white-space: nowrap; }

@media (max-width: 768px) {
  .client-edit-btn {
    width: 100%;
  }

  .client-edit-form__actions {
    flex-direction: column-reverse;
  }

  .client-edit-form__actions :deep(.app-button) {
    width: 100%;
  }
}
</style>
