<template>
  <div class="platform-page">
    <AppPageHeader :title="$t('platform.title')" :subtitle="$t('platform.subtitle')" />

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('platform.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="forbidden" padding="lg">
      <AppEmptyState icon="lock" :title="$t('platform.forbidden')" />
    </AppCard>

    <AppCard v-else-if="error" padding="lg">
      <AppEmptyState icon="error" :title="$t('platform.error')" />
    </AppCard>

    <template v-else-if="overview">
      <AppCard padding="lg" class="platform-page__llm">
        <h3 class="platform-page__section-title">{{ $t('platform.llm.title') }}</h3>
        <p class="platform-page__llm-desc">{{ $t('platform.llm.subtitle') }}</p>

        <AppCard v-if="settingsPending" padding="md">
          <p class="muted">{{ $t('platform.llm.loading') }}</p>
        </AppCard>

        <form v-else class="platform-page__llm-form" @submit.prevent="saveLlmSettings">
          <AppInput
            id="gemini-model"
            v-model="geminiModel"
            :label="$t('platform.llm.model_label')"
            list="gemini-model-suggestions"
            required
          />
          <p class="platform-page__llm-hint">{{ $t('platform.llm.model_hint') }}</p>
          <datalist id="gemini-model-suggestions">
            <option v-for="model in modelSuggestions" :key="model" :value="model" />
          </datalist>
          <p v-if="settingsSaveError" class="platform-page__llm-error" role="alert">
            {{ $t('platform.llm.save_error') }}
          </p>
          <p v-if="settingsSaved" class="platform-page__llm-success" role="status">
            {{ $t('platform.llm.save_success') }}
          </p>
          <div class="platform-page__llm-actions">
            <AppButton type="submit" variant="primary" size="sm" :disabled="settingsSaving">
              {{ settingsSaving ? $t('common.loading') : $t('common.save') }}
            </AppButton>
          </div>
        </form>
      </AppCard>

      <AppKpiGrid>
        <AppKpiCard
          icon="corporate_fare"
          tone="gold"
          :value="overview.summary.totalTenants"
          :label="$t('platform.kpi.total_tenants')"
        />
        <AppKpiCard
          icon="trending_up"
          tone="success"
          :value="overview.summary.activeTenants30d"
          :label="$t('platform.kpi.active_tenants')"
          :hint="$t('platform.kpi.active_hint')"
        />
        <AppKpiCard
          icon="group"
          tone="blue"
          :value="overview.summary.totalActiveUsers"
          :label="$t('platform.kpi.total_users')"
          :hint="seatsHint"
        />
        <AppKpiCard
          icon="payments"
          tone="warn"
          :value="trialCount"
          :label="$t('platform.kpi.trial_tenants')"
        />
      </AppKpiGrid>

      <AppListToolbar
        :filters="listFilters"
        :filter-values="filterValues"
        :sort-keys="sortKeys"
        :sort-key="sortKey"
        :sort-dir="sortDir"
        :has-active-filters="hasActiveFilters"
        @update:filter="setFilter"
        @update:sort-key="setSort($event)"
        @update:sort-dir="setSortDir"
        @reset="resetFilters"
      />

      <AppCard padding="lg" class="platform-page__table">
        <h3 class="platform-page__section-title">{{ $t('platform.tenants_title') }}</h3>
        <AppTable
          :columns="columns"
          :rows="displayRows"
          row-key="id"
          :empty-title="hasActiveFilters ? $t('common.list.no_results') : $t('platform.empty')"
        >
          <template #cell-societeName="{ value }">
            <span class="platform-page__tenant-name">{{ value }}</span>
          </template>
          <template #cell-subscriptionStatus="{ value }">
            <AppBadge :variant="statusVariant(value)">
              {{ statusLabel(value) }}
            </AppBadge>
          </template>
          <template #cell-activeUsers="{ row }">
            <span :class="{ 'platform-page__warn': seatWarn(row) }">
              {{ row.activeUsers }} / {{ row.seatLimit || '∞' }}
            </span>
          </template>
          <template #cell-activeLast30d="{ value }">
            <AppBadge :variant="value ? 'success' : 'default'">
              {{ value ? $t('platform.active_yes') : $t('platform.active_no') }}
            </AppBadge>
          </template>
          <template #cell-lastActivityAt="{ value }">
            {{ formatDate(value) }}
          </template>
        </AppTable>
      </AppCard>
    </template>
  </div>
</template>

<script setup lang="ts">
import { useListControls } from '~/composables/useListControls'

definePageMeta({
  middleware: ['platform']
})

const { t, locale } = useI18n()
const { overview, pending, error, forbidden, fetchOverview } = usePlatformOverview()
const {
  settings,
  pending: settingsPending,
  saving: settingsSaving,
  saveError: settingsSaveError,
  fetchSettings,
  saveSettings
} = usePlatformSettings()

const geminiModel = ref('')
const settingsSaved = ref(false)

const modelSuggestions = [
  'gemini-3.5-flash',
  'gemini-2.5-flash',
  'gemini-2.0-flash',
  'gemini-2.5-pro'
]

onMounted(async () => {
  await Promise.all([fetchOverview(), fetchSettings()])
  if (settings.value?.geminiModel) {
    geminiModel.value = settings.value.geminiModel
  }
})

watch(settings, (value) => {
  if (value?.geminiModel) {
    geminiModel.value = value.geminiModel
  }
})

async function saveLlmSettings() {
  settingsSaved.value = false
  const ok = await saveSettings(geminiModel.value.trim())
  if (ok) {
    settingsSaved.value = true
  }
}

const seatsHint = computed(() => {
  const limit = overview.value?.summary.totalSeatLimit ?? 0
  if (!limit) return undefined
  return t('platform.kpi.seats_hint', { n: limit })
})

const trialCount = computed(() => overview.value?.summary.tenantsByStatus?.trial ?? 0)

const listItems = computed(() =>
  (overview.value?.tenants ?? []).map((tenant) => ({
    ...tenant,
    societeName: tenant.societeName || tenant.name
  }))
)

const listFilters = computed(() => ({
  subscriptionStatus: {
    type: 'select' as const,
    label: t('platform.col.status'),
    options: ['trial', 'active', 'past_due', 'suspended', 'canceled', 'none'].map((status) => ({
      value: status,
      label: statusLabel(status)
    })),
    match: (row: { subscriptionStatus?: string }, value: string) =>
      String(row.subscriptionStatus ?? 'none') === value
  }
}))

const sortKeys = computed(() => [
  {
    key: 'societeName',
    label: t('platform.col.tenant'),
    type: 'string' as const,
    accessor: (row: { societeName?: string }) => row.societeName ?? ''
  },
  {
    key: 'lastActivityAt',
    label: t('platform.col.last_activity'),
    type: 'date' as const,
    accessor: (row: { lastActivityAt?: string }) => row.lastActivityAt ?? ''
  }
])

const {
  filterValues,
  sortKey,
  sortDir,
  sortedItems,
  hasActiveFilters,
  setFilter,
  setSort,
  setSortDir,
  resetFilters
} = useListControls(listItems, {
  storageKey: 'platform-tenants',
  defaultSort: { key: 'societeName', dir: 'asc' },
  filters: listFilters,
  sortKeys
})

const displayRows = computed(() => sortedItems.value)

const columns = computed(() => [
  { key: 'societeName', label: t('platform.col.tenant') },
  { key: 'subscriptionStatus', label: t('platform.col.status'), nowrap: true },
  { key: 'activeUsers', label: t('platform.col.seats'), nowrap: true },
  { key: 'modulesEnabled', label: t('platform.col.modules'), nowrap: true },
  { key: 'craCount', label: t('platform.col.cra'), nowrap: true },
  { key: 'tmaOpen', label: t('platform.col.tma_open'), nowrap: true },
  { key: 'budgetCount', label: t('platform.col.budget'), nowrap: true },
  { key: 'aiRequests30d', label: t('platform.col.ai'), nowrap: true },
  { key: 'activeLast30d', label: t('platform.col.active'), nowrap: true },
  { key: 'lastActivityAt', label: t('platform.col.last_activity'), nowrap: true }
])

function statusLabel(status: unknown): string {
  const key = String(status ?? 'none')
  if (key === 'none') return t('platform.status.none')
  const known = ['trial', 'active', 'past_due', 'suspended', 'canceled'] as const
  if ((known as readonly string[]).includes(key)) {
    return t(`dashboard.modules_panel.status.${key}`)
  }
  return key
}

function statusVariant(status: unknown): 'success' | 'warning' | 'default' {
  const key = String(status ?? 'none')
  switch (key) {
    case 'active':
      return 'success'
    case 'past_due':
    case 'suspended':
      return 'warning'
    default:
      return 'default'
  }
}

function seatWarn(row: Record<string, unknown>): boolean {
  const limit = Number(row.seatLimit ?? 0)
  const users = Number(row.activeUsers ?? 0)
  return limit > 0 && users / limit >= 0.9
}

function formatDate(value: unknown): string {
  if (!value || typeof value !== 'string') return '—'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return '—'
  return d.toLocaleDateString(locale.value === 'fr' ? 'fr-FR' : 'en-GB', {
    day: '2-digit',
    month: 'short',
    year: 'numeric'
  })
}
</script>

<style scoped>
.platform-page__section-title {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-lg);
  font-weight: 600;
}

.platform-page__llm {
  margin-bottom: var(--kore-space-md);
}

.platform-page__llm-desc {
  margin: 0 0 var(--kore-space-md);
  color: var(--kore-text-muted);
  font-size: var(--kore-text-sm);
}

.platform-page__llm-form {
  display: grid;
  gap: var(--kore-space-md);
  max-width: var(--kore-form-wide-max);
}

.platform-page__llm-hint {
  margin: calc(-1 * var(--kore-space-sm)) 0 0;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-sm);
}

.platform-page__llm-actions {
  display: flex;
  gap: var(--kore-space-sm);
}

.platform-page__llm-error {
  margin: 0;
  color: var(--kore-danger);
  font-size: var(--kore-text-sm);
}

.platform-page__llm-success {
  margin: 0;
  color: var(--kore-success);
  font-size: var(--kore-text-sm);
}

.platform-page__table {
  margin-top: var(--kore-space-md);
}

.platform-page__tenant-name {
  font-weight: 500;
}

.platform-page__warn {
  color: var(--kore-warn);
  font-weight: 600;
}

.muted {
  color: var(--kore-text-muted);
}
</style>
