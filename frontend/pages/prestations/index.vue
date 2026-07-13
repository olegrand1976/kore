<template>
  <div>
    <AppPageHeader :title="$t('prestations.title')">
      <template #actions>
        <AppButton variant="primary" size="sm" :disabled="validating" @click="validateAll">
          {{ $t('prestations.validate_all') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard padding="lg" class="prestations-filters">
      <label for="prestations-month">{{ $t('prestations.month') }}</label>
      <input id="prestations-month" v-model="month" type="month" @change="refresh">
    </AppCard>

    <AppCard v-if="pending" padding="lg">
      <CraSkeleton />
    </AppCard>

    <AppCard v-else-if="!rows.length" padding="lg">
      <AppEmptyState icon="schedule" :title="$t('prestations.empty')" />
    </AppCard>

    <AppCard v-else padding="none" class="prestations-table-wrap">
      <table class="prestations-table">
        <thead>
          <tr>
            <th>{{ $t('prestations.col_user') }}</th>
            <th>{{ $t('prestations.col_hours') }}</th>
            <th>{{ $t('prestations.col_weeks') }}</th>
            <th>{{ $t('prestations.col_status') }}</th>
            <th>{{ $t('prestations.col_reject_reason') }}</th>
            <th>{{ $t('prestations.col_actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in rows" :key="row.id">
            <td>{{ userLabel(row) }}</td>
            <td>{{ $t('cra.hours_value', { n: Math.round((row.totalMinutes ?? 0) / 60) }) }}</td>
            <td>
              <span>{{ weeksLabel(row) }}</span>
              <AppBadge v-if="hasAnomaly(row)" variant="warning" class="prestations-table__anomaly">
                {{ anomalyLabel(row) }}
              </AppBadge>
            </td>
            <td><AppBadge :variant="statusVariant(row.status)">{{ row.status }}</AppBadge></td>
            <td class="prestations-table__reason">{{ row.rejectReason?.trim() || '—' }}</td>
            <td class="prestations-table__actions">
              <AppButton variant="ghost" size="sm" @click="navigateTo(`/cra/${row.id}`)">
                {{ $t('prestations.open_cra') }}
              </AppButton>
              <AppButton
                v-if="row.status !== 'Définitif'"
                variant="secondary"
                size="sm"
                @click="rejectRow(row)"
              >
                {{ $t('prestations.reject') }}
              </AppButton>
            </td>
          </tr>
        </tbody>
      </table>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const { t } = useI18n()
const { statusVariant } = useCraStatus()

type PrestationRow = {
  id: string
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

const month = ref(new Date().toISOString().slice(0, 7))
const validating = ref(false)

const { data, pending, refresh } = await useFetch<{ data?: PrestationRow[] }>(
  () => `/api/prestations?month=${month.value}`
)

const rows = computed(() => data.value?.data ?? [])

const userLabel = (row: PrestationRow) => {
  const name = [row.userPrenom, row.userNom].filter(Boolean).join(' ').trim()
  return name || row.userLogin || '—'
}

const weeksLabel = (row: PrestationRow) => {
  const submitted = row.weeksSubmitted ?? 0
  const total = row.weeksTotal ?? submitted
  return t('prestations.weeks_ratio', { submitted, total })
}

const hasAnomaly = (row: PrestationRow) => {
  const submitted = row.weeksSubmitted ?? 0
  const total = row.weeksTotal ?? 0
  if (row.rejectReason?.trim() || row.rejectedAt) return true
  return total > 0 && submitted < total
}

const anomalyLabel = (row: PrestationRow) => {
  if (row.rejectReason?.trim() || row.rejectedAt) {
    return t('prestations.anomaly_rejected')
  }
  return t('prestations.anomaly_incomplete')
}

const validateAll = async () => {
  validating.value = true
  try {
    await $fetch('/api/prestations/validate-all', { method: 'POST', body: { month: month.value } })
    await refresh()
  } finally {
    validating.value = false
  }
}

const rejectRow = async (row: PrestationRow) => {
  const reason = window.prompt(t('prestations.reject_reason'))
  if (reason === null) return
  await $fetch(`/api/cra/timesheets/${row.id}/reject`, { method: 'POST', body: { reason } })
  await refresh()
}
</script>

<style scoped>
.prestations-filters {
  display: flex;
  align-items: center;
  gap: var(--kore-space-md);
  margin-bottom: var(--kore-space-lg);
}

.prestations-filters label {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

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

@media (max-width: 768px) {
  .prestations-table {
    font-size: var(--kore-text-small);
  }
}
</style>
