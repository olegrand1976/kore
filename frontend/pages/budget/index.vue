<template>
  <div>
    <AppPageHeader :title="$t('budget.title')" :subtitle="$t('budget.subtitle')" />

    <AppKpiGrid compact>
      <AppKpiCard
        icon="folder"
        tone="gold"
        :loading="pending"
        :value="kpi.total"
        :label="$t('budget.kpi_total')"
      />
      <AppKpiCard
        icon="event_available"
        tone="blue"
        :loading="pending"
        :value="kpi.plannedDays"
        :label="$t('budget.kpi_planned_days')"
      />
      <AppKpiCard
        icon="trending_up"
        tone="success"
        :loading="pending"
        :value="kpi.consumedDays"
        :label="$t('budget.kpi_consumed_days')"
        :hint="kpi.consumptionPct > 0 ? $t('budget.kpi_consumption_pct', { n: kpi.consumptionPct }) : undefined"
      />
      <AppKpiCard
        icon="warning"
        :tone="kpi.overrun > 0 ? 'warn' : 'success'"
        :loading="pending"
        :value="kpi.overrun"
        :label="$t('budget.kpi_overrun')"
        :hint="kpi.overrun > 0 ? $t('budget.kpi_overrun_hint', { n: kpi.overrun }) : undefined"
      />
    </AppKpiGrid>

    <AppCard v-if="loadError" padding="lg">
      <AppEmptyState icon="error" :title="loadError" />
    </AppCard>

    <template v-else>
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

      <AppCard padding="lg">
        <AppTable
          :columns="columns"
          :rows="displayRows"
          :loading="pending"
          :empty-title="hasActiveFilters ? $t('common.list.no_results') : $t('budget.empty')"
          :empty-description="hasActiveFilters ? undefined : $t('budget.empty_desc')"
        >
        <template #cell-application="{ row }">
          <button type="button" class="row-link" @click="navigateTo(`/budget/${row.id}`)">
            {{ row.application }}
          </button>
        </template>
        <template #cell-client="{ value }">
          <span :class="{ muted: !value }">{{ value || $t('budget.col_empty') }}</span>
        </template>
        <template #cell-type="{ row }">
          <AppBadge variant="gold">{{ row.typeLabel }}</AppBadge>
        </template>
        <template #cell-consumption="{ row }">
          <div class="consumption-cell">
            <div class="consumption-cell__track" role="progressbar" :aria-valuenow="row.consumptionPct" aria-valuemin="0" aria-valuemax="100">
              <div
                class="consumption-cell__fill"
                :class="`consumption-cell__fill--${row.status}`"
                :style="{ width: `${Math.min(100, row.consumptionPct)}%` }"
              />
            </div>
            <span class="consumption-cell__pct">{{ row.consumptionPct }} %</span>
            <AppBadge v-if="row.status === 'overrun'" variant="error">{{ $t('budget.status_overrun') }}</AppBadge>
          </div>
        </template>
        <template #cell-days="{ row }">
          {{ row.consumed }} / {{ row.planned }} {{ $t('budget.unit_days') }}
        </template>
        <template #cell-actions="{ row }">
          <AppButton variant="ghost" size="sm" @click="navigateTo(`/budget/${row.id}`)">
            {{ $t('budget.open') }}
          </AppButton>
        </template>
      </AppTable>
      </AppCard>
    </template>
  </div>
</template>

<script setup lang="ts">
import { budgetMetrics, consumptionPct } from '~/composables/useKpiMetrics'
import { useListControls } from '~/composables/useListControls'

definePageMeta({ layout: 'default' })

type BudgetRow = {
  id: string
  application: string
  client: string
  typeLabel: string
  type: string
  planned: number
  consumed: number
  consumptionPct: number
  status: string
}

const { t } = useI18n()
const { list, pickId, tripleValue } = useBudget()
const { list: listApplications, appById, pickAppLabel, pickAppClient } = useApplications()
const { budgetTypeLabel, budgetStatus, consumptionPercent } = useBudgetDisplay()
const { extractFetchError } = useApiError()

const loadError = ref('')

const { data, pending } = await useAsyncData('budget-list', async () => {
  loadError.value = ''
  try {
    const [budgets, applications] = await Promise.all([list(), listApplications()])
    return { budgets, applications }
  } catch (err) {
    loadError.value = extractFetchError(err)
    return { budgets: [], applications: [] }
  }
})

const appMap = computed(() => appById(data.value?.applications ?? []))

const kpi = computed(() => {
  const m = budgetMetrics(data.value?.budgets ?? [])
  return {
    total: m.total,
    plannedDays: m.plannedDays,
    consumedDays: m.consumedDays,
    overrun: m.overrun,
    consumptionPct: consumptionPct(m.consumedDays, m.plannedDays, false)
  }
})

const listItems = computed((): BudgetRow[] =>
  (data.value?.budgets ?? []).map((b) => {
    const id = pickId(b)
    const appId = b.applicationId ?? b.ApplicationID ?? ''
    const app = appMap.value.get(appId)
    const planned = tripleValue(b.planned ?? b.Planned, 'days')
    const consumed = tripleValue(b.consumed ?? b.Consumed, 'days')
    const type = b.type ?? b.Type ?? ''
    const status = budgetStatus(consumed, planned)
    return {
      id,
      application: pickAppLabel(app) || id.slice(0, 8),
      client: pickAppClient(app) || '',
      typeLabel: budgetTypeLabel(type),
      type,
      planned,
      consumed,
      consumptionPct: consumptionPercent(consumed, planned),
      status
    }
  })
)

const budgetTypes = computed(() => {
  const types = new Set<string>()
  for (const row of listItems.value) {
    if (row.type) types.add(row.type)
  }
  return [...types]
})

const listFilters = computed(() => ({
  type: {
    type: 'select' as const,
    label: t('budget.col_type'),
    options: budgetTypes.value.map((type) => ({
      value: type,
      label: budgetTypeLabel(type)
    })),
    match: (row: BudgetRow, value: string) => row.type === value
  },
  consumption: {
    type: 'select' as const,
    label: t('budget.col_consumption'),
    options: [
      { value: 'ok', label: t('budget.status_ok') },
      { value: 'warn', label: t('budget.status_warn') },
      { value: 'overrun', label: t('budget.status_overrun') }
    ],
    match: (row: BudgetRow, value: string) => row.status === value
  }
}))

const sortKeys = computed(() => [
  {
    key: 'consumptionPct',
    label: t('budget.col_consumption'),
    type: 'number' as const,
    accessor: (row: BudgetRow) => row.consumptionPct
  },
  {
    key: 'application',
    label: t('budget.col_application'),
    type: 'string' as const,
    accessor: (row: BudgetRow) => row.application
  },
  {
    key: 'client',
    label: t('budget.col_client'),
    type: 'string' as const,
    accessor: (row: BudgetRow) => row.client
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
  storageKey: 'budget-list',
  defaultSort: { key: 'consumptionPct', dir: 'desc' },
  filters: listFilters,
  sortKeys
})

const displayRows = computed(() => sortedItems.value)

const columns = computed(() => [
  { key: 'application', label: t('budget.col_application') },
  { key: 'client', label: t('budget.col_client') },
  { key: 'type', label: t('budget.col_type') },
  { key: 'consumption', label: t('budget.col_consumption') },
  { key: 'days', label: t('budget.col_days') },
  { key: 'actions', label: '' }
])
</script>

<style scoped>
.row-link {
  background: none;
  border: none;
  padding: 0;
  font: inherit;
  color: var(--kore-accent);
  cursor: pointer;
  text-align: left;
  text-decoration: underline;
  text-underline-offset: 2px;
}

.muted {
  color: var(--kore-text-muted);
}

.consumption-cell {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--kore-space-sm);
  min-width: 8rem;
}

.consumption-cell__track {
  flex: 1 1 4rem;
  height: 0.4rem;
  background: var(--kore-bg-subtle);
  border-radius: var(--kore-radius-full);
  overflow: hidden;
}

.consumption-cell__fill {
  height: 100%;
  border-radius: var(--kore-radius-full);
}

.consumption-cell__fill--ok {
  background: var(--kore-accent);
}

.consumption-cell__fill--warn {
  background: var(--kore-brand-gold);
}

.consumption-cell__fill--overrun {
  background: var(--kore-danger);
}

.consumption-cell__pct {
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
  white-space: nowrap;
}
</style>
