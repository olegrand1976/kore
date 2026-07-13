<template>
  <div>
    <AppKpiGrid compact class="conges-kpis">
      <AppKpiCard
        icon="list_alt"
        tone="gold"
        :loading="pending"
        :value="kpi.total"
        :label="$t('conges.kpi_total')"
      />
      <AppKpiCard
        icon="hourglass_top"
        tone="warn"
        :loading="pending"
        :value="kpi.pending"
        :label="$t('conges.kpi_pending')"
      />
      <AppKpiCard
        icon="check_circle"
        tone="success"
        :loading="pending"
        :value="kpi.approved"
        :label="$t('conges.kpi_approved')"
      />
      <AppKpiCard
        icon="cancel"
        tone="blue"
        :loading="pending"
        :value="kpi.rejected"
        :label="$t('conges.kpi_rejected')"
      />
      <AppKpiCard
        v-for="(balance, idx) in kpi.balanceKpis"
        :key="balance.code"
        :icon="idx === 0 ? 'beach_access' : 'free_cancellation'"
        :tone="idx === 0 ? 'gold' : 'blue'"
        :loading="balancesPending"
        :value="formatBalance(balance.remaining)"
        :label="balance.label"
        :hint="balanceHint(balance.taken, balance.acquired)"
      />
      <AppKpiCard
        icon="event_upcoming"
        tone="success"
        :loading="pending"
        :value="kpi.upcomingDays"
        :label="$t('conges.kpi_upcoming_days')"
        :hint="$t('conges.kpi_upcoming_hint')"
      />
      <AppKpiCard
        icon="date_range"
        tone="blue"
        :loading="pending"
        :value="kpi.approvedDays"
        :label="$t('conges.kpi_approved_days')"
      />
    </AppKpiGrid>

    <div class="conges-overview">
      <AppCard padding="lg" class="conges-overview__panel">
        <h3 class="conges-panel__title">{{ $t('conges.chart_title') }}</h3>
        <p class="conges-panel__desc">{{ $t('conges.chart_desc') }}</p>
        <AppBarChart
          :bars="statusBars"
          :loading="pending"
          :empty-label="$t('conges.chart_empty')"
        />
      </AppCard>

      <AppCard padding="lg" class="conges-overview__panel">
        <h3 class="conges-panel__title">{{ $t('conges.balances_summary_title') }}</h3>
        <p class="conges-panel__desc">{{ $t('conges.balances_summary_desc') }}</p>
        <AppTable
          :columns="balanceColumns"
          :rows="balanceRows"
          :loading="balancesPending"
          :empty-title="$t('conges.balances_empty')"
        />
      </AppCard>
    </div>

    <AppCard v-if="showForm" padding="lg" class="conges-form-card">
      <h3 class="conges-panel__title">{{ $t('conges.form_title') }}</h3>
      <form class="conges-form" @submit.prevent="submitRequest">
        <div class="conges-form__select">
          <label for="leave-type">{{ $t('conges.col_type') }}</label>
          <select id="leave-type" v-model="form.type" required>
            <option v-for="lt in activeLeaveTypes" :key="pickLeaveTypeCode(lt)" :value="pickLeaveTypeCode(lt)">
              {{ pickLeaveTypeLabel(lt) }}
            </option>
          </select>
        </div>
        <AppInput id="from" v-model="form.from" type="date" :label="$t('conges.from')" required />
        <AppInput id="to" v-model="form.to" type="date" :label="$t('conges.to')" required />
        <AppInput id="motif" v-model="form.motif" :label="$t('conges.motif')" required />
        <div class="conges-form__actions">
          <AppButton variant="ghost" size="sm" type="button" @click="showForm = false">
            {{ $t('common.cancel') }}
          </AppButton>
          <AppButton variant="primary" size="sm" type="submit" :disabled="submitting">
            {{ $t('conges.submit') }}
          </AppButton>
        </div>
      </form>
    </AppCard>

    <AppCard v-if="pending" padding="lg">
      <p class="conges-muted">{{ $t('conges.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="rows.length" padding="none" class="conges-table-wrap">
      <AppTable :columns="columns" :rows="rows" row-key="id">
        <template #cell-type="{ value }">
          <span class="conges-type">{{ typeLabel(String(value)) }}</span>
        </template>
        <template #cell-from="{ value }">
          <span class="conges-date">{{ formatDate(String(value)) }}</span>
        </template>
        <template #cell-to="{ value }">
          <span class="conges-date">{{ formatDate(String(value)) }}</span>
        </template>
        <template #cell-days="{ value }">
          <span class="conges-days">{{ $t('conges.days_value', { n: value }) }}</span>
        </template>
        <template #cell-motif="{ value }">
          <span class="conges-motif" :title="String(value)">{{ value || $t('conges.motif_empty') }}</span>
        </template>
        <template #cell-status="{ value }">
          <AppBadge :variant="statusVariant(String(value))">{{ statusLabel(String(value)) }}</AppBadge>
        </template>
        <template #cell-decidedAt="{ value }">
          <span class="conges-date conges-date--muted">{{ formatDate(String(value)) }}</span>
        </template>
      </AppTable>
    </AppCard>

    <AppCard v-else padding="lg">
      <AppEmptyState icon="event_busy" :title="$t('conges.empty')" :description="$t('conges.empty_desc')">
        <AppButton variant="primary" size="sm" @click="showForm = true">{{ $t('conges.new_request') }}</AppButton>
      </AppEmptyState>
    </AppCard>

    <p v-if="errorMsg" class="conges-flash conges-flash--error" role="alert">{{ errorMsg }}</p>
  </div>
</template>

<script setup lang="ts">
import { leaveMetrics, leaveStatusSeries } from '~/composables/useKpiMetrics'
import {
  pickLeaveTypeCode,
  pickLeaveTypeLabel,
  useLeave,
  useLeaveLabels,
  useLeaveTypeConfigs
} from '~/composables/useLeave'

const { t, locale } = useI18n()
const { extractFetchError } = useApiError()
const {
  list,
  create,
  pickStatus,
  pickType,
  pickMotif,
  pickFrom,
  pickTo,
  pickDecidedAt,
  leaveDayCount,
  balances
} = useLeave()
const { fetchMine, activeTypes } = useLeaveTypeConfigs()
const { typeLabel, statusLabel, statusVariant } = useLeaveLabels()

const { data, pending, refresh } = await useAsyncData('leave-requests', () => list())
const { data: balancesData, pending: balancesPending, refresh: refreshBalances } = await useAsyncData(
  'leave-balances',
  () => balances()
)
const { data: leaveTypesData } = await useAsyncData('leave-types-mine', () => fetchMine())

const showForm = ref(false)
const submitting = ref(false)
const errorMsg = ref('')

const indexActions = useState<{ toggleForm: () => void } | null>('conges-index-actions', () => null)
indexActions.value = {
  toggleForm: () => {
    showForm.value = !showForm.value
  }
}
onUnmounted(() => {
  indexActions.value = null
})

const form = reactive<{ type: string; from: string; to: string; motif: string }>({
  type: '',
  from: '',
  to: '',
  motif: ''
})

const activeLeaveTypes = computed(() => activeTypes.value)
const leaveTypes = computed(() => leaveTypesData.value ?? [])

watch(
  activeLeaveTypes,
  (types) => {
    if (!form.type && types.length > 0) {
      form.type = pickLeaveTypeCode(types[0])
    }
  },
  { immediate: true }
)

const items = computed(() => data.value ?? [])
const balanceItems = computed(() => balancesData.value ?? [])

const kpi = computed(() => leaveMetrics(items.value, balanceItems.value, leaveTypes.value))

const statusBars = computed(() =>
  leaveStatusSeries(items.value, (status) => statusLabel(status))
)

const balanceColumns = computed(() => [
  { key: 'type', label: t('conges.col_type') },
  { key: 'acquired', label: t('conges.balances_acquired') },
  { key: 'taken', label: t('conges.balances_taken') },
  { key: 'remaining', label: t('conges.balances_amount') }
])

const balanceRows = computed(() =>
  balanceItems.value.map((item) => ({
    type: typeLabel(item.type ?? item.Type ?? ''),
    acquired: formatBalance(item.acquired ?? item.Acquired),
    taken: formatBalance(item.taken ?? item.Taken),
    remaining: formatBalance(item.remaining ?? item.Remaining ?? item.balance ?? item.Balance)
  }))
)

const columns = computed(() => [
  { key: 'type', label: t('conges.col_type') },
  { key: 'from', label: t('conges.from') },
  { key: 'to', label: t('conges.to') },
  { key: 'days', label: t('conges.col_days') },
  { key: 'motif', label: t('conges.motif') },
  { key: 'status', label: t('conges.col_status') },
  { key: 'decidedAt', label: t('conges.col_decided') }
])

const rows = computed(() =>
  items.value.map((item) => {
    const from = pickFrom(item)
    const to = pickTo(item)
    const decidedAt = pickDecidedAt(item)
    return {
      id: item.id ?? item.ID ?? `${from}-${to}`,
      type: pickType(item),
      from,
      to,
      days: leaveDayCount(from, to),
      motif: pickMotif(item),
      status: pickStatus(item),
      decidedAt: decidedAt ? decidedAt.slice(0, 10) : '—'
    }
  })
)

const formatBalance = (value: number | null | undefined) => {
  if (value == null || Number.isNaN(value)) return '—'
  return Number.isInteger(value) ? String(value) : value.toFixed(1)
}

const balanceHint = (taken: number, acquired: number) => {
  if (acquired <= 0 && taken <= 0) return undefined
  return t('conges.balance_hint', { taken, acquired })
}

const formatDate = (raw: string) => {
  if (!raw || raw === '—') return '—'
  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) return '—'
  return date.toLocaleDateString(locale.value === 'en' ? 'en-US' : 'fr-FR', {
    day: 'numeric',
    month: 'short',
    year: 'numeric'
  })
}

const submitRequest = async () => {
  submitting.value = true
  errorMsg.value = ''
  try {
    await create({ type: form.type, from: form.from, to: form.to, motif: form.motif })
    showForm.value = false
    form.from = ''
    form.to = ''
    form.motif = ''
    form.type = activeLeaveTypes.value[0] ? pickLeaveTypeCode(activeLeaveTypes.value[0]) : ''
    await Promise.all([refresh(), refreshBalances()])
  } catch (err) {
    errorMsg.value = extractFetchError(err)
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.conges-kpis :deep(.kpi-card) {
  height: 100%;
}

@media (min-width: 640px) {
  .conges-kpis {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

@media (min-width: 1100px) {
  .conges-kpis {
    grid-template-columns: repeat(8, minmax(0, 1fr));
  }
}

.conges-overview {
  display: grid;
  gap: var(--kore-space-md);
  margin-bottom: var(--kore-space-lg);
}

@media (min-width: 900px) {
  .conges-overview {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

.conges-panel__title {
  margin: 0 0 var(--kore-space-xs);
  font-size: var(--kore-text-h3);
  font-weight: 600;
  color: var(--kore-text);
}

.conges-panel__desc {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.conges-form-card {
  margin-bottom: var(--kore-space-lg);
}

.conges-form {
  display: grid;
  gap: var(--kore-space-md);
  max-width: var(--kore-form-max);
}

.conges-form__select {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}

.conges-form__select label {
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
  font-weight: 500;
}

.conges-form__select select {
  background: var(--kore-bg-elevated);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  color: var(--kore-text);
  font-family: var(--kore-font);
  font-size: var(--kore-text-body);
  padding: 0.75rem 1rem;
}

.conges-form__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
}

.conges-table-wrap {
  overflow: hidden;
}

.conges-type {
  font-weight: 500;
  color: var(--kore-text);
}

.conges-date {
  white-space: nowrap;
  color: var(--kore-text);
}

.conges-date--muted {
  color: var(--kore-text-muted);
  font-size: var(--kore-text-caption);
}

.conges-days {
  font-variant-numeric: tabular-nums;
  color: var(--kore-text);
}

.conges-motif {
  display: inline-block;
  max-width: 14rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: bottom;
  color: var(--kore-text);
}

.conges-muted {
  margin: 0;
  color: var(--kore-text-muted);
}

.conges-flash {
  margin-top: var(--kore-space-md);
  font-size: var(--kore-text-small);
}

.conges-flash--error {
  color: var(--kore-error);
}

@media (max-width: 768px) {
  .conges-form__actions :deep(.app-btn) {
    flex: 1 1 calc(50% - var(--kore-space-sm));
  }

  .conges-motif {
    max-width: 8rem;
  }
}
</style>
