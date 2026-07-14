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

    <AppCard padding="lg" class="prestations-filters">
      <label for="prestations-month">{{ $t('prestations.month') }}</label>
      <input id="prestations-month" v-model="month" type="month" @change="refresh">
    </AppCard>

    <p v-if="actionMsg" class="flash" :class="{ 'flash--error': actionError }" role="status">{{ actionMsg }}</p>

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
definePageMeta({ layout: 'default' })

const { t } = useI18n()
const { statusVariant } = useCraStatus()

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

const ettAlertByUser = computed(() => {
  const map = new Map<string, boolean>()
  for (const report of ettData.value?.data ?? []) {
    if (report.userId) {
      map.set(String(report.userId), Boolean(report.alert))
    }
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
  if (!draft?.status || draft.status === 'created') {
    return draft?.status === 'created' ? t('prestations.invoice_created') : t('prestations.validate_ok')
  }
  return t('prestations.invoice_skipped', { reason: draft.reason ?? 'unknown' })
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
