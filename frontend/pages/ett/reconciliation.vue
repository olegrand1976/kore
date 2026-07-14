<template>
  <div>
    <AppPageHeader :title="$t('ett.reconciliation_title')" />

    <AppListToolbar
      v-if="canValidateEtt && !pending"
      :filters="listFilters"
      :filter-values="filterValues"
      :sort-keys="sortKeys"
      :sort-key="sortKey"
      :sort-dir="sortDir"
      :has-active-filters="hasActiveFilters"
      @update:filter="onFilterUpdate"
      @update:sort-key="setSort($event)"
      @update:sort-dir="setSortDir"
      @reset="onResetFilters"
    />

    <AppCard v-else padding="lg" class="ett-filters">
      <label for="ett-month">{{ $t('prestations.month') }}</label>
      <input id="ett-month" v-model="month" type="month" @change="refresh">
    </AppCard>

    <AppCard v-if="pending" padding="lg">
      <CraSkeleton />
    </AppCard>

    <AppCard v-else-if="errorMsg" padding="lg">
      <p class="flash flash--error">{{ errorMsg }}</p>
    </AppCard>

    <AppCard v-else-if="displayTeamReports.length > 0" padding="lg" class="ett-team">
      <div class="table-wrap">
        <table class="ett-team__table">
          <thead>
            <tr>
              <th>{{ $t('prestations.col_user') }}</th>
              <th>{{ $t('ett.cra_hours') }}</th>
              <th>{{ $t('ett.ett_hours') }}</th>
              <th>{{ $t('ett.delta_hours') }}</th>
              <th>{{ $t('ett.missing_ett_days') }}</th>
              <th>{{ $t('prestations.col_status') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in displayTeamReports" :key="row.userId">
              <td>{{ row.userName || row.userLogin || row.userId }}</td>
              <td>{{ formatHours(row.craHours) }}</td>
              <td>{{ formatHours(row.ettHours) }}</td>
              <td>{{ formatHours(row.deltaHours) }}</td>
              <td>{{ row.missingEttDays ?? 0 }}</td>
              <td>
                <AppBadge :variant="row.alert ? 'warning' : 'success'">
                  {{ row.alert ? $t('ett.alert') : $t('ett.ok') }}
                </AppBadge>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </AppCard>

    <AppCard v-else-if="report" padding="lg" class="ett-report">
      <AppBadge :variant="report.alert ? 'warning' : 'success'">
        {{ report.alert ? $t('ett.alert') : $t('ett.ok') }}
      </AppBadge>
      <dl class="ett-report__grid">
        <div>
          <dt>{{ $t('ett.cra_hours') }}</dt>
          <dd>{{ formatHours(report.craHours) }}</dd>
        </div>
        <div>
          <dt>{{ $t('ett.ett_hours') }}</dt>
          <dd>{{ formatHours(report.ettHours) }}</dd>
        </div>
        <div>
          <dt>{{ $t('ett.delta_hours') }}</dt>
          <dd>{{ formatHours(report.deltaHours) }}</dd>
        </div>
        <div>
          <dt>{{ $t('ett.missing_ett_days') }}</dt>
          <dd>{{ report.missingEttDays ?? 0 }}</dd>
        </div>
      </dl>
      <p v-if="report.alertMessage" class="ett-report__message">{{ report.alertMessage }}</p>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
import { syncListMonthFilter, useListControls } from '~/composables/useListControls'

definePageMeta({ layout: 'default' })

const { t } = useI18n()
const { canValidateEtt } = usePermissions()

type ReconciliationReport = {
  userId?: string
  userLogin?: string
  userName?: string
  craHours?: number
  ettHours?: number
  deltaHours?: number
  missingEttDays?: number
  alert?: boolean
  alertMessage?: string
}

const month = ref(new Date().toISOString().slice(0, 7))
const errorMsg = ref('')

const endpoint = computed(() => {
  const params = new URLSearchParams({ month: month.value })
  if (canValidateEtt.value) {
    params.set('scope', 'team')
  }
  return `/api/ett/reconciliation?${params.toString()}`
})

const { data, pending, refresh, error } = await useFetch<{ data?: ReconciliationReport | ReconciliationReport[] }>(endpoint)

watch(error, (err) => {
  errorMsg.value = err ? t('ett.load_error') : ''
})

const teamReports = computed(() => {
  const payload = data.value?.data
  if (canValidateEtt.value && Array.isArray(payload)) {
    return payload
  }
  return []
})

type EttTeamRow = ReconciliationReport & { userId: string }

const listItems = computed((): EttTeamRow[] =>
  teamReports.value.filter((row): row is EttTeamRow => Boolean(row.userId)).map((row) => ({
    ...row,
    userId: String(row.userId)
  }))
)

const listFilters = computed(() => ({
  month: {
    type: 'month' as const,
    label: t('prestations.month'),
    defaultValue: month.value,
    match: (_row: EttTeamRow, _value: string) => true
  },
  alert: {
    type: 'select' as const,
    label: t('prestations.col_status'),
    options: [
      { value: 'true', label: t('ett.alert') },
      { value: 'false', label: t('ett.ok') }
    ],
    match: (row: EttTeamRow, value: string) => String(Boolean(row.alert)) === value
  }
}))

const sortKeys = computed(() => [
  {
    key: 'deltaHours',
    label: t('ett.delta_hours'),
    type: 'number' as const,
    accessor: (row: EttTeamRow) => Math.abs(Number(row.deltaHours ?? 0))
  },
  {
    key: 'userName',
    label: t('prestations.col_user'),
    type: 'string' as const,
    accessor: (row: EttTeamRow) => row.userName || row.userLogin || row.userId
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
  storageKey: 'ett-reconciliation',
  defaultSort: { key: 'deltaHours', dir: 'desc' },
  filters: listFilters,
  sortKeys
})

const displayTeamReports = computed(() => sortedItems.value)

const onFilterUpdate = (key: string, value: string) => {
  setFilter(key, value)
  if (key === 'month' && value && value !== month.value) {
    month.value = value
    refresh()
  }
}

const onResetFilters = () => {
  resetFilters()
  if (filterValues.month && filterValues.month !== month.value) {
    month.value = filterValues.month
    refresh()
  }
}

watch(month, (next) => {
  if (filterValues.month !== next) {
    filterValues.month = next
  }
})

syncListMonthFilter(filterValues, month, refresh)

const report = computed(() => {
  if (canValidateEtt.value) return null
  const payload = data.value?.data
  return payload && !Array.isArray(payload) ? payload : null
})

const formatHours = (value?: number) => {
  if (value == null) return '—'
  return t('cra.hours_value', { n: Math.round(value * 10) / 10 })
}
</script>

<style scoped>
.ett-filters {
  display: flex;
  align-items: center;
  gap: var(--kore-space-md);
  margin-bottom: var(--kore-space-lg);
}

.ett-filters label {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.ett-report__grid {
  display: grid;
  gap: var(--kore-space-md);
  margin-top: var(--kore-space-lg);
}

@media (min-width: 769px) {
  .ett-report__grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

.ett-report__grid dt {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.ett-report__grid dd {
  margin: 0;
  font-size: var(--kore-text-h3);
  font-weight: 600;
}

.ett-report__message {
  margin-top: var(--kore-space-md);
  color: var(--kore-text-muted);
}

.ett-team__table {
  width: 100%;
  border-collapse: collapse;
}

.ett-team__table th,
.ett-team__table td {
  padding: 0.625rem 0.75rem;
  border-bottom: 1px solid var(--kore-border);
  text-align: left;
  font-size: var(--kore-text-small);
}

.table-wrap {
  overflow-x: auto;
}

.flash--error {
  color: var(--kore-error);
}
</style>
