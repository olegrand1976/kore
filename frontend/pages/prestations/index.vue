<template>
  <div>
    <AppPageHeader :title="$t('prestations.title')">
      <template #actions>
        <AppButton variant="secondary" size="sm" @click="exportXml">
          {{ $t('prestations.export_xml') }}
        </AppButton>
        <AppButton variant="primary" size="sm" :disabled="validating" @click="validateAll">
          {{ $t('prestations.validate_all') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppListToolbar
      v-if="!pending"
      :filters="listFilters"
      :filter-values="filterValues"
      :sort-keys="sortKeys"
      :sort-key="sortKey"
      :sort-dir="sortDir"
      :view="view"
      kanban-enabled
      :has-active-filters="hasActiveFilters"
      @update:filter="onFilterUpdate"
      @update:sort-key="setSort($event)"
      @update:sort-dir="setSortDir"
      @update:view="setView"
      @reset="onResetFilters"
    />

    <p v-if="actionMsg" class="flash" :class="{ 'flash--error': actionError }" role="status">{{ actionMsg }}</p>

    <AppCard v-if="pending" padding="lg">
      <CraSkeleton />
    </AppCard>

    <AppCard v-else-if="!displayRows.length" padding="lg">
      <AppEmptyState
        icon="schedule"
        :title="hasActiveFilters ? $t('common.list.no_results') : $t('prestations.empty')"
      />
    </AppCard>

    <AppCard v-else-if="view === 'table'" padding="none" class="prestations-table-wrap">
      <table class="prestations-table">
        <thead>
          <tr>
            <th>{{ $t('prestations.col_user') }}</th>
            <th>{{ $t('prestations.col_hours') }}</th>
            <th>{{ $t('prestations.col_weeks') }}</th>
            <th>{{ $t('prestations.col_ett') }}</th>
            <th>{{ $t('prestations.col_status') }}</th>
            <th>{{ $t('prestations.col_reject_reason') }}</th>
            <th>{{ $t('prestations.col_actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in displayRows" :key="row.id">
            <td>{{ userLabel(row) }}</td>
            <td>{{ $t('cra.hours_value', { n: Math.round((row.totalMinutes ?? 0) / 60) }) }}</td>
            <td>
              <span>{{ weeksLabel(row) }}</span>
              <AppBadge v-if="hasAnomaly(row)" variant="warning" class="prestations-table__anomaly">
                {{ anomalyLabel(row) }}
              </AppBadge>
            </td>
            <td class="prestations-table__ett">
              <span>{{ ettDeltaLabel(row) }}</span>
              <AppButton
                v-if="hasEttAlert(row)"
                variant="ghost"
                size="sm"
                @click="openEttReconciliation(row)"
              >
                {{ $t('prestations.ett_reconcile') }}
              </AppButton>
            </td>
            <td><AppBadge :variant="statusVariant(row.status)">{{ statusLabel(row.status) }}</AppBadge></td>
            <td class="prestations-table__reason">{{ row.rejectReason?.trim() || '—' }}</td>
            <td class="prestations-table__actions">
              <AppButton variant="ghost" size="sm" @click="navigateTo(`/cra/${row.id}`)">
                {{ $t('prestations.open_cra') }}
              </AppButton>
              <AppButton variant="ghost" size="sm" @click="downloadPdf(row)">
                {{ $t('prestations.download_pdf') }}
              </AppButton>
              <AppButton
                v-if="row.status === 'ValidéSemaine'"
                variant="primary"
                size="sm"
                :disabled="rowActionId === row.id"
                @click="validateRow(row)"
              >
                {{ $t('prestations.validate') }}
              </AppButton>
              <AppButton
                v-if="row.status !== 'Définitif'"
                variant="secondary"
                size="sm"
                @click="openReject(row)"
              >
                {{ $t('prestations.reject') }}
              </AppButton>
            </td>
          </tr>
        </tbody>
      </table>
    </AppCard>

    <AppCard v-else padding="lg">
      <AppKanbanBoard
        :columns="kanbanColumns"
        :items="displayRows"
        :column-key="(row) => String((row as PrestationRow).status)"
        :item-key="(row) => String((row as PrestationRow).id)"
        :empty-label="$t('common.list.no_results')"
      >
        <template #card="{ item }">
          <div class="prestations-kanban-card">
            <p class="prestations-kanban-card__user">{{ userLabel(item as PrestationRow) }}</p>
            <p class="prestations-kanban-card__meta">
              {{ $t('cra.hours_value', { n: Math.round(((item as PrestationRow).totalMinutes ?? 0) / 60) }) }}
              · {{ weeksLabel(item as PrestationRow) }}
            </p>
            <p v-if="ettDeltaLabel(item as PrestationRow) !== '—'" class="prestations-kanban-card__ett">
              {{ ettDeltaLabel(item as PrestationRow) }}
            </p>
            <div v-if="hasAnomaly(item as PrestationRow)" class="prestations-kanban-card__badges">
              <AppBadge variant="warning">{{ anomalyLabel(item as PrestationRow) }}</AppBadge>
            </div>
            <AppBadge :variant="statusVariant((item as PrestationRow).status)">
              {{ statusLabel((item as PrestationRow).status) }}
            </AppBadge>
            <div class="prestations-kanban-card__actions">
              <AppButton variant="ghost" size="sm" @click="navigateTo(`/cra/${(item as PrestationRow).id}`)">
                {{ $t('prestations.open_cra') }}
              </AppButton>
              <AppButton variant="ghost" size="sm" @click="downloadPdf(item as PrestationRow)">
                {{ $t('prestations.download_pdf') }}
              </AppButton>
              <AppButton
                v-if="(item as PrestationRow).status === 'ValidéSemaine'"
                variant="primary"
                size="sm"
                :disabled="rowActionId === (item as PrestationRow).id"
                @click="validateRow(item as PrestationRow)"
              >
                {{ $t('prestations.validate') }}
              </AppButton>
              <AppButton
                v-if="(item as PrestationRow).status !== 'Définitif'"
                variant="secondary"
                size="sm"
                @click="openReject(item as PrestationRow)"
              >
                {{ $t('prestations.reject') }}
              </AppButton>
            </div>
          </div>
        </template>
      </AppKanbanBoard>
    </AppCard>

    <AppModal v-model:open="rejectOpen" width="md" :title-id="rejectTitleId" :aria-label="$t('prestations.reject')">
      <form class="reject-form" @submit.prevent="confirmReject">
        <label :for="rejectReasonId">{{ $t('prestations.reject_reason') }}</label>
        <textarea :id="rejectReasonId" v-model="rejectReason" rows="3" required />
        <div class="reject-form__actions">
          <AppButton variant="ghost" size="sm" type="button" @click="rejectOpen = false">
            {{ $t('common.cancel') }}
          </AppButton>
          <AppButton variant="primary" size="sm" type="submit" :disabled="rejecting">
            {{ $t('prestations.reject') }}
          </AppButton>
        </div>
      </form>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
import { syncListMonthFilter, useListControls } from '~/composables/useListControls'

definePageMeta({ layout: 'default' })

const CRA_STATUSES = ['Brouillon', 'ValidéSemaine', 'Définitif'] as const

const { t } = useI18n()
const { statusVariant, statusLabel } = useCraStatus()
const { mapInvoiceDraftMessage } = useCraError()

type PrestationRow = {
  id: string
  userId?: string
  userPrenom?: string
  userNom?: string
  userLogin?: string
  totalMinutes?: number
  weeksSubmitted?: number
  weeksTotal?: number
  rejectReason?: string
  rejectedAt?: string
  status: string
}

type EttReport = {
  userId?: string
  alert?: boolean
  deltaHours?: number
  ettHours?: number
  craHours?: number
}

type ValidateResponse = {
  data?: {
    invoiceDraft?: { status?: string; reason?: string }
  }
}

type ValidateAllResponse = {
  data?: {
    validated?: number
    failed?: Array<{ timesheetId?: string; reason?: string }>
  }
}

const month = ref(new Date().toISOString().slice(0, 7))
const validating = ref(false)
const rowActionId = ref('')
const actionMsg = ref('')
const actionError = ref(false)
const rejectOpen = ref(false)
const rejectReason = ref('')
const rejecting = ref(false)
const rejectTarget = ref<PrestationRow | null>(null)
const rejectTitleId = 'prestations-reject-title'
const rejectReasonId = 'prestations-reject-reason'

const { data, pending, refresh } = await useFetch<{ data?: PrestationRow[] }>(
  () => `/api/prestations?month=${month.value}`
)

const { data: ettData } = await useFetch<{ data?: EttReport[] }>(
  () => `/api/ett/reconciliation?month=${month.value}&scope=team`,
  { watch: [month] }
)

const rows = computed(() => data.value?.data ?? [])

const listFilters = computed(() => ({
  month: {
    type: 'month' as const,
    label: t('prestations.month'),
    defaultValue: month.value,
    match: (_row: PrestationRow, _value: string) => true
  },
  status: {
    type: 'select' as const,
    label: t('prestations.col_status'),
    options: CRA_STATUSES.map((status) => ({
      value: status,
      label: statusLabel(status)
    })),
    match: (row: PrestationRow, value: string) => row.status === value
  }
}))

const sortKeys = computed(() => [
  {
    key: 'user',
    label: t('prestations.col_user'),
    type: 'string' as const,
    accessor: (row: PrestationRow) => userLabel(row)
  },
  {
    key: 'hours',
    label: t('prestations.col_hours'),
    type: 'number' as const,
    accessor: (row: PrestationRow) => row.totalMinutes ?? 0
  },
  {
    key: 'status',
    label: t('prestations.col_status'),
    type: 'string' as const,
    accessor: (row: PrestationRow) => row.status
  }
])

const {
  filterValues,
  sortKey,
  sortDir,
  view,
  sortedItems,
  hasActiveFilters,
  setFilter,
  setSort,
  setSortDir,
  setView,
  resetFilters
} = useListControls(rows, {
  storageKey: 'prestations-list',
  defaultSort: { key: 'user', dir: 'asc' },
  kanbanEnabled: true,
  filters: listFilters,
  sortKeys
})

const displayRows = computed(() => sortedItems.value)

const kanbanColumns = computed(() =>
  CRA_STATUSES.map((status) => ({
    id: status,
    label: statusLabel(status),
    tone: status === 'Définitif' ? 'success' as const : status === 'ValidéSemaine' ? 'warn' as const : 'muted' as const
  }))
)

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

const ettReportByUser = computed(() => {
  const map = new Map<string, EttReport>()
  for (const report of ettData.value?.data ?? []) {
    if (report.userId) {
      map.set(String(report.userId), report)
    }
  }
  return map
})

const ettAlertByUser = computed(() => {
  const map = new Map<string, boolean>()
  for (const [userId, report] of ettReportByUser.value.entries()) {
    map.set(userId, Boolean(report.alert))
  }
  return map
})

const userLabel = (row: PrestationRow) => {
  const name = [row.userPrenom, row.userNom].filter(Boolean).join(' ').trim()
  return name || row.userLogin || '—'
}

const weeksLabel = (row: PrestationRow) => {
  const submitted = row.weeksSubmitted ?? 0
  const total = row.weeksTotal ?? submitted
  return t('prestations.weeks_ratio', { submitted, total })
}

const hasEttAlert = (row: PrestationRow) =>
  Boolean(row.userId && ettAlertByUser.value.get(String(row.userId)))

const ettDeltaLabel = (row: PrestationRow) => {
  if (!row.userId) return '—'
  const report = ettReportByUser.value.get(String(row.userId))
  if (!report) return '—'
  const delta = Number(report.deltaHours ?? 0)
  if (Math.abs(delta) < 0.05) return t('prestations.ett_ok')
  return t('prestations.ett_delta', { delta: delta.toFixed(1) })
}

const openEttReconciliation = (row: PrestationRow) => {
  const query = new URLSearchParams({ month: month.value })
  if (row.userId) query.set('userId', row.userId)
  navigateTo(`/ett/reconciliation?${query.toString()}`)
}

const hasAnomaly = (row: PrestationRow) => {
  const submitted = row.weeksSubmitted ?? 0
  const total = row.weeksTotal ?? 0
  if (row.rejectReason?.trim() || row.rejectedAt) return true
  if (row.userId && ettAlertByUser.value.get(String(row.userId))) return true
  return total > 0 && submitted < total
}

const anomalyLabel = (row: PrestationRow) => {
  if (row.userId && ettAlertByUser.value.get(String(row.userId))) {
    return t('prestations.anomaly_ett')
  }
  if (row.rejectReason?.trim() || row.rejectedAt) {
    return t('prestations.anomaly_rejected')
  }
  return t('prestations.anomaly_incomplete')
}

const invoiceDraftMessage = (draft?: { status?: string; reason?: string }) => {
  if (!draft?.status) return t('prestations.validate_ok')
  if (draft.status === 'created') return t('prestations.invoice_created')
  if (draft.status === 'unavailable') return t('prestations.invoice_unavailable')
  return mapInvoiceDraftMessage(draft, 'prestations.invoice_skipped')
}

const setActionMsg = (msg: string, isError = false) => {
  actionMsg.value = msg
  actionError.value = isError
}

const validateAll = async () => {
  validating.value = true
  setActionMsg('')
  try {
    const res = await $fetch<ValidateAllResponse>('/api/prestations/validate-all', {
      method: 'POST',
      body: { month: month.value }
    })
    const validated = res?.data?.validated ?? 0
    const failed = res?.data?.failed?.length ?? 0
    setActionMsg(t('prestations.validate_all_result', { validated, failed }), failed > 0 && validated === 0)
    await refresh()
  } catch {
    setActionMsg(t('prestations.action_error'), true)
  } finally {
    validating.value = false
  }
}

const validateRow = async (row: PrestationRow) => {
  rowActionId.value = row.id
  setActionMsg('')
  try {
    const res = await $fetch<ValidateResponse>(`/api/cra/timesheets/${row.id}/validate`, { method: 'POST' })
    setActionMsg(invoiceDraftMessage(res?.data?.invoiceDraft))
    await refresh()
  } catch {
    setActionMsg(t('prestations.action_error'), true)
  } finally {
    rowActionId.value = ''
  }
}

const downloadPdf = async (row: PrestationRow) => {
  setActionMsg('')
  try {
    const res = await $fetch<Blob>(`/api/cra/timesheets/${row.id}/pdf`, { method: 'POST', responseType: 'blob' })
    const url = URL.createObjectURL(res)
    const link = document.createElement('a')
    link.href = url
    link.download = `cra-${month.value}-${row.userLogin ?? row.id}.pdf`
    link.click()
    URL.revokeObjectURL(url)
  } catch {
    setActionMsg(t('prestations.pdf_error'), true)
  }
}

const exportXml = () => {
  window.open(`/api/prestations/export.xml?month=${encodeURIComponent(month.value)}`, '_blank')
}

const openReject = (row: PrestationRow) => {
  rejectTarget.value = row
  rejectReason.value = ''
  rejectOpen.value = true
}

const confirmReject = async () => {
  if (!rejectTarget.value) return
  rejecting.value = true
  setActionMsg('')
  try {
    await $fetch(`/api/cra/timesheets/${rejectTarget.value.id}/reject`, {
      method: 'POST',
      body: { reason: rejectReason.value.trim() }
    })
    rejectOpen.value = false
    setActionMsg(t('prestations.reject_ok'))
    await refresh()
  } catch {
    setActionMsg(t('prestations.action_error'), true)
  } finally {
    rejecting.value = false
  }
}
</script>

<style scoped>
.prestations-table-wrap {
  overflow-x: auto;
}

.prestations-table {
  width: 100%;
  border-collapse: collapse;
}

.prestations-table th,
.prestations-table td {
  padding: var(--kore-space-md);
  text-align: left;
  border-bottom: 1px solid var(--kore-border);
}

.prestations-table__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-xs);
}

.prestations-table__anomaly {
  display: inline-block;
  margin-top: var(--kore-space-xs);
}

.prestations-table__reason {
  max-width: 12rem;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.reject-form {
  display: grid;
  gap: var(--kore-space-md);
}

.reject-form textarea {
  width: 100%;
  padding: 0.625rem 0.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  color: var(--kore-text);
}

.reject-form__actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--kore-space-sm);
}

.flash {
  margin-bottom: var(--kore-space-md);
  font-size: var(--kore-text-small);
  color: var(--kore-success);
}

.flash--error {
  color: var(--kore-error);
}

.prestations-kanban-card {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-sm);
}

.prestations-kanban-card__user {
  margin: 0;
  font-weight: 600;
  font-size: var(--kore-text-small);
  color: var(--kore-text);
}

.prestations-kanban-card__meta {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.prestations-kanban-card__ett {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.prestations-kanban-card__badges {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-xs);
}

.prestations-kanban-card__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-xs);
}

@media (max-width: 768px) {
  .prestations-table thead {
    display: none;
  }

  .prestations-table tr {
    display: grid;
    gap: var(--kore-space-xs);
    padding: var(--kore-space-md);
    border-bottom: 1px solid var(--kore-border);
  }

  .prestations-table td {
    border: none;
    padding: 0;
  }

  .prestations-table__actions {
    margin-top: var(--kore-space-sm);
  }
}
</style>
