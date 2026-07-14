<template>
  <div>
    <AppPageHeader :title="displayName" :subtitle="$t('fiche.collaborateur_title')">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/cra')">
          <AppIcon name="arrow_back" /> {{ $t('fiche.back_cra') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('fiche.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="error" padding="lg">
      <AppEmptyState icon="error" :title="$t('fiche.not_found')" />
    </AppCard>

    <template v-else-if="user">
      <AppCard padding="lg" class="fiche-hero">
        <div class="fiche-hero__main">
          <div class="fiche-hero__avatar" aria-hidden="true">
            <AppIcon name="person" />
          </div>
          <div>
            <h2 class="fiche-hero__name">{{ displayName }}</h2>
            <p class="fiche-hero__login">{{ user.login }}</p>
            <div class="fiche-hero__badges">
              <AppBadge variant="default">{{ user.profil }}</AppBadge>
              <AppBadge :variant="user.active ? 'success' : 'default'">
                {{ user.active ? $t('users.active') : $t('users.inactive') }}
              </AppBadge>
            </div>
          </div>
        </div>
      </AppCard>

      <div class="fiche-grid">
        <AppCard padding="lg">
          <h3 class="fiche-section-title">{{ $t('fiche.section_account') }}</h3>
          <dl class="fiche-dl">
            <div>
              <dt>{{ $t('fiche.col_login') }}</dt>
              <dd>{{ user.login || $t('fiche.none') }}</dd>
            </div>
            <div>
              <dt>{{ $t('fiche.col_email') }}</dt>
              <dd>{{ user.email || $t('fiche.none') }}</dd>
            </div>
            <div>
              <dt>{{ $t('fiche.col_account_type') }}</dt>
              <dd>{{ accountTypeLabel(user.typeCompte) }}</dd>
            </div>
            <div>
              <dt>{{ $t('fiche.col_lang') }}</dt>
              <dd>{{ user.langue?.toUpperCase() || $t('fiche.none') }}</dd>
            </div>
            <div>
              <dt>{{ $t('fiche.col_cra_required') }}</dt>
              <dd>{{ user.craRequis ? $t('fiche.yes') : $t('fiche.no') }}</dd>
            </div>
            <div v-if="user.salarieETT">
              <dt>{{ $t('fiche.col_ett') }}</dt>
              <dd>{{ $t('fiche.yes') }}</dd>
            </div>
          </dl>
        </AppCard>

        <AppCard padding="lg">
          <h3 class="fiche-section-title">{{ $t('fiche.section_org') }}</h3>
          <dl class="fiche-dl">
            <div>
              <dt>{{ $t('fiche.col_team') }}</dt>
              <dd>{{ user.equipeLibelle || $t('fiche.none') }}</dd>
            </div>
            <div>
              <dt>{{ $t('fiche.col_profile') }}</dt>
              <dd>{{ user.profil || $t('fiche.none') }}</dd>
            </div>
          </dl>
        </AppCard>

        <AppCard padding="lg">
          <h3 class="fiche-section-title">{{ $t('fiche.section_validity') }}</h3>
          <dl class="fiche-dl">
            <div>
              <dt>{{ $t('fiche.col_activation') }}</dt>
              <dd>{{ formatDate(user.dateActivation) }}</dd>
            </div>
            <div>
              <dt>{{ $t('fiche.col_expiration') }}</dt>
              <dd>{{ user.dateExpiration ? formatDate(user.dateExpiration) : $t('fiche.none') }}</dd>
            </div>
          </dl>
        </AppCard>
      </div>

      <AppCard v-if="showCraSection" padding="none" class="fiche-table-wrap">
        <div class="fiche-table-head">
          <h3 class="fiche-section-title">{{ $t('fiche.section_cra') }}</h3>
        </div>
        <div class="fiche-table-toolbar">
          <AppListToolbar
            :filters="craListFilters"
            :filter-values="craFilterValues"
            :sort-keys="craSortKeys"
            :sort-key="craSortKey"
            :sort-dir="craSortDir"
            :has-active-filters="craHasActiveFilters"
            @update:filter="setCraFilter"
            @update:sort-key="setCraSort($event)"
            @update:sort-dir="setCraSortDir"
            @reset="resetCraFilters"
          />
        </div>
        <AppTable
          :columns="craColumns"
          :rows="craDisplayRows"
          row-key="id"
          :empty-title="craHasActiveFilters ? $t('common.list.no_results') : $t('fiche.cra_empty')"
        >
          <template #cell-month="{ value }">
            <span class="fiche-strong">{{ formatMonth(String(value)) }}</span>
          </template>
          <template #cell-status="{ value }">
            <AppBadge :variant="statusVariant(String(value))">{{ statusLabel(String(value)) }}</AppBadge>
          </template>
          <template #cell-client="{ value }">
            <span class="fiche-truncate">{{ value || $t('cra.context_empty') }}</span>
          </template>
          <template #cell-actions="{ row }">
            <AppButton variant="ghost" size="sm" @click="navigateTo(`/cra/${row.id}`)">
              {{ $t('fiche.open_cra') }}
            </AppButton>
          </template>
        </AppTable>
      </AppCard>
    </template>
  </div>
</template>

<script setup lang="ts">
import { formatUserDisplayName } from '~/composables/useUserDisplay'
import { useListControls } from '~/composables/useListControls'

definePageMeta({ layout: 'default' })

const CRA_STATUSES = ['Brouillon', 'ValidéSemaine', 'Définitif'] as const

type UserDetail = {
  id?: string
  login?: string
  prenom?: string
  nom?: string
  email?: string
  profil?: string
  active?: boolean
  langue?: string
  typeCompte?: string
  craRequis?: boolean
  salarieETT?: boolean
  equipeLibelle?: string
  dateActivation?: string
  dateExpiration?: string | null
}

type CraRow = {
  id: string
  userId?: string
  month: string
  status: string
  commercialInfo?: { client?: string; mission?: string }
}

const route = useRoute()
const { t, locale } = useI18n()
const { formatDate } = useFicheFormat()
const { statusLabel, statusVariant } = useCraStatus()
const { canValidateCra } = usePermissions()
const { user: sessionUser, fetchSession } = useAuth()

await fetchSession()

const id = computed(() => String(route.params.id ?? ''))

const { data, pending, error } = await useFetch<UserDetail>(() => `/api/org/users/${id.value}`, {
  watch: [id]
})

const { data: craData } = await useFetch('/api/cra/timesheets/recent', {
  lazy: true,
  server: false
})

const user = computed(() => {
  const payload = (data.value as { data?: UserDetail })?.data ?? data.value
  return payload && typeof payload === 'object' ? payload : null
})

const displayName = computed(() =>
  formatUserDisplayName(user.value?.prenom, user.value?.nom, user.value?.login)
)

const showCraSection = computed(
  () => canValidateCra.value || sessionUser.value?.userId === id.value
)

const craItems = computed((): CraRow[] => {
  if (!showCraSection.value) return []
  const payload = (craData.value as { data?: CraRow[] })?.data ?? craData.value
  if (!Array.isArray(payload)) return []
  return payload.filter((ts) => String(ts.userId ?? '') === id.value)
})

const craListFilters = computed(() => ({
  status: {
    type: 'select' as const,
    label: t('cra.col_status'),
    options: CRA_STATUSES.map((status) => ({
      value: status,
      label: statusLabel(status)
    })),
    match: (row: { status: string }, value: string) => row.status === value
  }
}))

const craSortKeys = computed(() => [
  { key: 'month', label: t('cra.col_period'), type: 'date' as const, accessor: (row: { month: string }) => row.month }
])

const {
  filterValues: craFilterValues,
  sortKey: craSortKey,
  sortDir: craSortDir,
  sortedItems: craSortedItems,
  hasActiveFilters: craHasActiveFilters,
  setFilter: setCraFilter,
  setSort: setCraSort,
  setSortDir: setCraSortDir,
  resetFilters: resetCraFilters
} = useListControls(craItems, {
  storageKey: 'collaborateur-cra',
  defaultSort: { key: 'month', dir: 'desc' },
  filters: craListFilters,
  sortKeys: craSortKeys
})

const craColumns = computed(() => [
  { key: 'month', label: t('cra.col_period') },
  { key: 'client', label: t('cra.col_client') },
  { key: 'status', label: t('cra.col_status') },
  { key: 'actions', label: '' }
])

const craDisplayRows = computed(() =>
  craSortedItems.value.map((ts) => ({
    id: ts.id,
    month: ts.month,
    client: ts.commercialInfo?.client ?? '',
    status: ts.status,
    actions: ''
  }))
)

const accountTypeLabel = (type?: string) => {
  switch (type) {
    case 'Interne':
      return t('fiche.account_type_interne')
    case 'Client':
      return t('fiche.account_type_client')
    case 'Prestataire':
      return t('fiche.account_type_prestataire')
    default:
      return type || t('fiche.none')
  }
}

const formatMonth = (raw: string) => {
  const [y, m] = raw.split('-').map(Number)
  if (!y || !m) return raw
  return new Date(y, m - 1, 1).toLocaleDateString(locale.value === 'en' ? 'en-US' : 'fr-FR', {
    month: 'long',
    year: 'numeric'
  })
}
</script>

<style scoped>
.muted { color: var(--kore-text-muted); }

.fiche-hero { margin-bottom: var(--kore-space-lg); }

.fiche-hero__main {
  display: flex;
  align-items: center;
  gap: var(--kore-space-lg);
}

.fiche-hero__avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 4rem;
  height: 4rem;
  border-radius: var(--kore-radius-full);
  background: var(--kore-bg-subtle);
  border: 1px solid var(--kore-border);
  color: var(--kore-brand-gold);
  font-size: 2rem;
}

.fiche-hero__name {
  margin: 0 0 var(--kore-space-xs);
  font-size: var(--kore-text-h2);
}

.fiche-hero__login {
  margin: 0 0 var(--kore-space-sm);
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

.fiche-hero__badges {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-xs);
}

.fiche-grid {
  display: grid;
  gap: var(--kore-space-lg);
  margin-bottom: var(--kore-space-lg);
}

@media (min-width: 768px) {
  .fiche-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

.fiche-section-title {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-h3);
}

.fiche-dl {
  display: grid;
  gap: var(--kore-space-md);
  margin: 0;
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

.fiche-truncate {
  max-width: 12rem;
  display: inline-block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: bottom;
}

@media (max-width: 768px) {
  .fiche-hero__main {
    flex-direction: column;
    align-items: flex-start;
  }

  .fiche-truncate {
    max-width: 8rem;
  }
}
</style>
