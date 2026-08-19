<template>
  <div>
    <AppPageHeader :title="pageTitle">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/cra')">
          <AppIcon name="arrow_back" /> {{ $t('cra.back') }}
        </AppButton>
        <AppButton
          v-if="canValidateCra"
          variant="secondary"
          size="sm"
          :disabled="!canEdit || saving"
          @click="onValidateFinal"
        >
          {{ $t('cra.validate_final') }}
        </AppButton>
        <AppButton
          v-if="canValidateCra && timesheet?.status !== 'Définitif'"
          variant="secondary"
          size="sm"
          @click="rejectOpen = true"
        >
          {{ $t('cra.reject') }}
        </AppButton>
        <CraDeleteTimesheetControl
          v-if="timesheet"
          :timesheet-id="id"
          :status="timesheet.status"
          :month="timesheet.month"
          @changed="onTimesheetAdminChange"
          @error="onTimesheetDeleteError"
        />
        <AppButton
          v-if="canEdit"
          variant="secondary"
          size="sm"
          :disabled="prefillLoading"
          @click="loadPrefillETT"
        >
          {{ $t('cra.prefill_ett') }}
        </AppButton>
        <AppButton
          v-if="canEdit"
          variant="secondary"
          size="sm"
          :disabled="prefillLoading"
          @click="loadPrefillHolidays"
        >
          {{ $t('cra.prefill_holidays') }}
        </AppButton>
        <AppButton
          v-if="canEdit"
          variant="secondary"
          size="sm"
          :disabled="prefillLoading"
          @click="loadPrefillSuggest"
        >
          {{ $t('ai.cra_prefill') }}
        </AppButton>
        <AppButton
          variant="secondary"
          size="sm"
          :disabled="downloading || !canDownload"
          :aria-describedby="canDownload ? undefined : 'cra-download-hint'"
          @click="openPdfPreview"
        >
          {{ $t('cra.preview_pdf') }}
        </AppButton>
        <AppButton
          variant="primary"
          size="sm"
          :disabled="downloading || !canDownload"
          :aria-describedby="canDownload ? undefined : 'cra-download-hint'"
          @click="downloadPdf"
        >
          <AppIcon name="download" /> {{ $t('cra.download') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard v-if="loading" padding="lg">
      <CraSkeleton />
    </AppCard>

    <AppCard v-else-if="error" padding="lg">
      <AppEmptyState icon="error" :title="$t('cra.not_found')" />
    </AppCard>

    <div v-else-if="timesheet" class="cra-detail">
      <AppCard v-if="saving" padding="lg" class="cra-detail__saving">
        <CraSkeleton />
      </AppCard>
      <AppCard padding="lg" class="cra-detail__meta">
        <dl class="meta">
          <div><dt>{{ $t('cra.period') }}</dt><dd>{{ formatMonth(timesheet.month) }}</dd></div>
          <div>
            <dt>{{ $t('cra.col_status') }}</dt>
            <dd><AppBadge :variant="statusVariant(timesheet.status)">{{ statusLabel(timesheet.status) }}</AppBadge></dd>
          </div>
          <div>
            <dt>{{ $t('cra.client') }}</dt>
            <dd>
              <NuxtLink
                v-if="prestation.clientId"
                :to="`/clients/${prestation.clientId}`"
                class="meta__link"
              >
                {{ prestation.client || $t('cra.context_empty') }}
              </NuxtLink>
              <span
                v-else
                :class="{ 'meta__empty': !prestation.client }"
              >
                {{ prestation.client || $t('cra.context_empty') }}
              </span>
            </dd>
          </div>
          <div>
            <dt>{{ $t('cra.mission') }}</dt>
            <dd>
              <NuxtLink
                v-if="prestation.missionId"
                :to="`/missions/${prestation.missionId}`"
                class="meta__link"
              >
                {{ prestation.mission || $t('cra.context_empty') }}
              </NuxtLink>
              <span
                v-else
                :class="{ 'meta__empty': !prestation.mission }"
              >
                {{ prestation.mission || $t('cra.context_empty') }}
              </span>
            </dd>
          </div>
        </dl>
        <CraMonthlyPreview
          class="cra-detail__preview"
          :total-minutes="totalMinutes"
          :capacity-minutes="capacityMinutes"
          :weeks-submitted="weeksSubmitted"
          :weeks-total="weeksTotal"
          :prefill-ratio="prefillRatio"
          :progress="progress"
        />
      </AppCard>

      <div class="cra-detail__body">
        <aside class="cra-detail__aside">
          <PrestationInfoForm
            ref="prestationFormRef"
            :client="prestation.client"
            :mission="prestation.mission"
            :client-id="prestation.clientId"
            :mission-id="prestation.missionId"
            :missions="missions"
            :description="prestation.description"
            :technologies="prestation.technologies"
            :lieu="prestation.lieu"
            :responsable-client="prestation.responsableClient"
            :disabled="!canEdit"
            :saving="savingPrestation"
            :message="prestationMsg"
            :is-error="prestationError"
            @change="onPrestationChange"
            @submit="savePrestation"
          />
        </aside>

        <div class="cra-detail__main">
          <TimesheetGrid
            :weeks="selectedWeeks"
            :month="timesheet.month"
            :week-start-day="weekStartDay"
            :day-capacity-minutes="dayCapacityMinutes"
            :week-submit-policy="weekSubmitPolicy"
            :can-edit="canEdit"
            :saving="saving"
            :missions="missions"
            :task-types="taskTypesEnabled"
            :work-ref-options="workRefOptions"
            :work-ref-label-for="workRefLabelFor"
            @save="onSaveWeek"
            @submit="onSubmitWeek"
          />

          <AppCard v-if="anomalies.length || anomaliesLoading" padding="lg" class="cra-detail__anomalies">
            <h3 class="cra-detail__anomalies-title">{{ $t('cra.anomalies_title') }}</h3>
            <p v-if="anomaliesLoading" class="muted">{{ $t('cra.loading') }}</p>
            <ul v-else-if="anomalies.length" class="cra-detail__anomalies-list">
              <li v-for="(item, idx) in anomalies" :key="idx">{{ item }}</li>
            </ul>
            <p v-else class="muted">{{ $t('cra.anomalies_empty') }}</p>
          </AppCard>
        </div>
      </div>
    </div>

    <CraPdfPreview
      v-model:open="pdfPreviewOpen"
      :loading="pdfPreviewLoading"
      :error="pdfPreviewError"
      :preview-url="pdfPreviewUrl"
      @download="downloadPdf"
    />

    <p v-if="validateMsg" class="flash" role="status">
      {{ validateMsg }}
      <NuxtLink v-if="invoiceLink" :to="invoiceLink" class="flash__link">{{ $t('cra.invoice_created_link') }}</NuxtLink>
    </p>
    <p v-if="prefillMsg" class="flash flash--info" role="status">{{ prefillMsg }}</p>
    <p v-if="downloadError" class="flash flash--error" role="alert">{{ downloadError }}</p>
    <p v-if="actionError" class="flash flash--error" role="alert">{{ actionError }}</p>

    <AppModal v-model:open="rejectOpen" width="md" :title-id="rejectTitleId" :aria-label="$t('cra.reject')">
      <form class="reject-form" @submit.prevent="confirmReject">
        <label :for="rejectReasonId">{{ $t('cra.reject_reason') }}</label>
        <textarea :id="rejectReasonId" v-model="rejectReason" rows="3" required />
        <div class="reject-form__actions">
          <AppButton variant="ghost" size="sm" type="button" @click="rejectOpen = false">
            {{ $t('common.cancel') }}
          </AppButton>
          <AppButton variant="primary" size="sm" type="submit" :disabled="rejecting">
            {{ $t('cra.reject') }}
          </AppButton>
        </div>
      </form>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
import type { CraLine } from '~/stores/cra'
import { weekNumberForDay } from '~/composables/useWeekCalendar'
import { useCraMonthStats } from '~/composables/useCraMonthStats'
import { useCraWorkRefs } from '~/composables/useCraWorkRefs'
import { prestationInfoComplete, type PrestationInfoFields } from '~/utils/craPrestation'

definePageMeta({ layout: 'default' })

const { apiFetch } = useApiFetch()
const route = useRoute()
const { t, locale } = useI18n()
const { statusLabel, statusVariant } = useCraStatus()
const { canValidateCra } = usePermissions()
const { mapCraError, mapInvoiceDraftMessage: mapInvoiceDraft } = useCraError()
const id = computed(() => String(route.params.id))

const { timesheet, loading, error, canEdit, selectedWeeks, saving, load, saveWeek, submitWeek, validateFinal, rejectTimesheet } = useCra(id)
const { user, fetchSession } = useAuth()
const { options: workRefOptions, load: loadWorkRefs, labelFor: workRefLabelFor } = useCraWorkRefs()

const weekStartDay = ref(1)
const dayCapacityMinutes = ref(480)
const weekSubmitPolicy = ref<'block' | 'warn' | 'none'>('warn')
const taskTypesEnabled = ref<string[]>(['manual', 'interne', 'formation', 'mission'])
const missions = ref<Array<{ id: string; clientName?: string; clientId?: string; label?: string }>>([])
const prestationFormRef = ref<{ local: Record<string, unknown> } | null>(null)
const pdfPreviewOpen = ref(false)
const pdfPreviewLoading = ref(false)
const pdfPreviewError = ref('')
const pdfPreviewUrl = ref('')

const loadOrgSettings = async () => {
  try {
    const res = await apiFetch<{
      data?: {
        weekStartDay?: number
        dayCapacityMinutes?: number
        weekSubmitPolicy?: string
        taskTypesEnabled?: string[]
      }
      weekStartDay?: number
      dayCapacityMinutes?: number
      weekSubmitPolicy?: string
      taskTypesEnabled?: string[]
    }>('/api/org/users/me/calendar-settings')
    const data = res.data ?? res
    const day = data.weekStartDay
    if (day != null) {
      const normalizedDay = Number(day)
      if (Number.isFinite(normalizedDay) && normalizedDay >= 0 && normalizedDay <= 6) {
        weekStartDay.value = normalizedDay
      }
    }
    if (data.dayCapacityMinutes != null) {
      const normalizedCapacity = Number(data.dayCapacityMinutes)
      if (Number.isFinite(normalizedCapacity) && normalizedCapacity > 0) {
        dayCapacityMinutes.value = normalizedCapacity
      }
    }
    const policy = data.weekSubmitPolicy
    if (policy === 'block' || policy === 'warn' || policy === 'none') {
      weekSubmitPolicy.value = policy
    }
    if (Array.isArray(data.taskTypesEnabled) && data.taskTypesEnabled.length > 0) {
      taskTypesEnabled.value = data.taskTypesEnabled
    }
  } catch {
    weekStartDay.value = 1
    dayCapacityMinutes.value = 480
    weekSubmitPolicy.value = 'warn'
    taskTypesEnabled.value = ['manual', 'interne', 'formation', 'mission']
  }
}

const loadMissions = async () => {
  try {
    const res = await apiFetch<{ data: Array<Record<string, unknown>> }>('/api/ssii/missions')
    missions.value = (res.data ?? []).map((item) => {
      const clientName = String(item.clientName ?? item.ClientName ?? '')
      const startDate = String(item.startDate ?? item.StartDate ?? '').slice(0, 10)
      return {
        id: String(item.id ?? item.ID ?? ''),
        clientName,
        clientId: String(item.clientId ?? item.ClientID ?? ''),
        label: startDate ? `${clientName || 'Mission'} (${startDate})` : clientName
      }
    }).filter((m) => m.id)
  } catch {
    missions.value = []
  }
}

const savingPrestation = ref(false)
const prestationMsg = ref('')
const prestationError = ref(false)
const downloading = ref(false)
const downloadError = ref('')
const prefillLoading = ref(false)
const prefillMsg = ref('')
const actionError = ref('')
const validateMsg = ref('')
const invoiceLink = ref('')
const rejectOpen = ref(false)
const rejectReason = ref('')
const rejecting = ref(false)
const rejectTitleId = 'cra-reject-title'
const rejectReasonId = 'cra-reject-reason'
const anomalies = ref<string[]>([])
const anomaliesLoading = ref(false)
const { suggestCraPrefill, fetchCraAnomalies } = useAi()

const loadAnomalies = async () => {
  anomaliesLoading.value = true
  try {
    const res = await fetchCraAnomalies(id.value) as { data?: { anomalies?: string[] }; anomalies?: string[] }
    const list = res?.data?.anomalies ?? res?.anomalies ?? []
    anomalies.value = Array.isArray(list) ? list.map(String) : []
  } catch {
    anomalies.value = []
  } finally {
    anomaliesLoading.value = false
  }
}

const loadPrefillETT = async () => {
  if (!timesheet.value) return
  prefillLoading.value = true
  prefillMsg.value = ''
  actionError.value = ''
  try {
    const res = await apiFetch<{ data?: { added?: number } }>(`/api/cra/timesheets/${id.value}/prefill-ett`, {
      method: 'POST'
    })
    await load(id.value)
    await loadAnomalies()
    prefillMsg.value = t('cra.prefill_ett_result', { n: res?.data?.added ?? 0 })
  } catch (err) {
    actionError.value = mapCraError(err)
  } finally {
    prefillLoading.value = false
  }
}

await Promise.all([load(id.value), loadOrgSettings(), loadMissions(), fetchSession()])
const userId = user.value?.userId ?? ''
if (userId) {
  await loadWorkRefs(userId)
}
await loadAnomalies()

const monthRef = computed(() => timesheet.value?.month ?? '')
const weekStartDayRef = computed(() => weekStartDay.value)
const weeksRef = computed(() => selectedWeeks.value)
const {
  totalMinutes,
  capacityMinutes,
  weeksSubmitted,
  weeksTotal,
  prefillRatio,
  progress
} = useCraMonthStats(weeksRef, monthRef, weekStartDayRef, dayCapacityMinutes)

const prestation = reactive({
  client: '',
  mission: '',
  clientId: '' as string,
  missionId: '' as string,
  description: '',
  technologies: [] as string[],
  lieu: '',
  responsableClient: ''
})

const mergePrefillLines = (existing: CraLine[], suggestions: Array<{ day: string; duration: number; comment?: string }>): CraLine[] => {
  const result = existing.map((line) => ({ ...line }))
  for (const suggestion of suggestions) {
    const day = suggestion.day.slice(0, 10)
    const hasManual = result.some(
      (line) =>
        line.day.slice(0, 10) === day &&
        line.duration > 0 &&
        (line.origin === 'manual' || (line.sourceType === 'manual' && line.sourceId !== 'default'))
    )
    if (hasManual) continue

    const duration = Math.round(suggestion.duration * 60)
    if (duration <= 0) continue

    const idx = result.findIndex(
      (line) => line.day.slice(0, 10) === day && line.sourceType === 'manual' && line.sourceId === 'default'
    )
    const line: CraLine = {
      sourceType: 'manual',
      sourceId: 'default',
      day,
      duration,
      comment: suggestion.comment ?? '',
      origin: 'prefill'
    }
    if (idx >= 0) {
      result[idx] = line
    } else {
      result.push(line)
    }
  }
  return result
}

const loadPrefillHolidays = async () => {
  if (!timesheet.value) return
  prefillLoading.value = true
  prefillMsg.value = ''
  try {
    const res = await apiFetch<{ data?: { added?: number } }>(`/api/cra/timesheets/${id.value}/prefill-holidays`, {
      method: 'POST'
    })
    await load(id.value)
    prefillMsg.value = t('cra.prefill_holidays_result', { n: res?.data?.added ?? 0 })
  } catch (err) {
    actionError.value = mapCraError(err)
  } finally {
    prefillLoading.value = false
  }
}

const loadPrefillSuggest = async () => {
  if (!timesheet.value) return
  prefillLoading.value = true
  prefillMsg.value = ''
  try {
    const res = await suggestCraPrefill(id.value)
    if (res.lines.length === 0) {
      prefillMsg.value = t('ai.cra_prefill_result', { n: 0 })
      return
    }

    const byWeek = new Map<number, CraLine[]>()
    for (const suggestion of res.lines) {
      const weekNumber = weekNumberForDay(timesheet.value.month, suggestion.day, weekStartDay.value)
      const week = selectedWeeks.value.find((w) => w.weekNumber === weekNumber)
      const current = byWeek.get(weekNumber) ?? week?.lines.map((line) => ({ ...line })) ?? []
      byWeek.set(weekNumber, mergePrefillLines(current, [suggestion]))
    }

    for (const [weekNumber, lines] of byWeek) {
      await saveWeek(weekNumber, lines)
    }
    await load(id.value)
    prefillMsg.value = t('ai.cra_prefill_result', { n: res.lines.length })
  } catch (err) {
    actionError.value = mapCraError(err)
  } finally {
    prefillLoading.value = false
  }
}

watch(timesheet, (ts) => {
  if (!ts?.commercialInfo) return
  prestation.client = ts.commercialInfo.client ?? ''
  prestation.mission = ts.commercialInfo.mission ?? ''
  prestation.clientId = ts.commercialInfo.clientId ?? ''
  prestation.missionId = ts.commercialInfo.missionId ?? ''
  prestation.description = ts.commercialInfo.description ?? ''
  prestation.technologies = [...(ts.commercialInfo.technologies ?? [])]
  prestation.lieu = ts.commercialInfo.lieu ?? ''
  prestation.responsableClient = ts.commercialInfo.responsableClient ?? ''
}, { immediate: true })

const canDownload = computed(() => prestationInfoComplete(prestation.client, prestation.mission))

const pageTitle = computed(() => {
  if (!timesheet.value?.month) return t('cra.title')
  return t('cra.detail_title', { period: formatMonth(timesheet.value.month) })
})

const onTimesheetAdminChange = async (action: 'unvalidate' | 'delete') => {
  switch (action) {
    case 'delete':
      await navigateTo('/cra')
      return
    case 'unvalidate':
      validateMsg.value = t('cra.unvalidate_ok')
      await load()
      return
    default: {
      const _exhaustive: never = action
      return _exhaustive
    }
  }
}

const onTimesheetDeleteError = (message: string) => {
  actionError.value = message
}

const formatMonth = (raw: string) => {
  const [y, m] = raw.split('-').map(Number)
  return new Date(y, m - 1, 1).toLocaleDateString(locale.value === 'en' ? 'en-US' : 'fr-FR', {
    month: 'long', year: 'numeric'
  })
}

const onSaveWeek = async (weekNumber: number, lines: Parameters<typeof saveWeek>[1]) => {
  actionError.value = ''
  try {
    await saveWeek(weekNumber, lines)
    await loadAnomalies()
  } catch (err) {
    actionError.value = mapCraError(err)
  }
}

const onSubmitWeek = async (weekNumber: number) => {
  actionError.value = ''
  try {
    await submitWeek(weekNumber)
    await loadAnomalies()
  } catch (err) {
    actionError.value = mapCraError(err)
  }
}

const onValidateFinal = async () => {
  actionError.value = ''
  validateMsg.value = ''
  invoiceLink.value = ''
  try {
    await persistPrestation()
    const draft = await validateFinal()
    validateMsg.value = mapInvoiceDraft(draft)
    const invoiceId = (draft as { invoiceId?: string } | undefined)?.invoiceId
    if (draft?.status === 'created' && invoiceId) {
      invoiceLink.value = `/facturation/${invoiceId}`
    }
    await loadAnomalies()
  } catch (err) {
    actionError.value = mapCraError(err)
  }
}

const confirmReject = async () => {
  if (!rejectReason.value.trim()) return
  rejecting.value = true
  actionError.value = ''
  try {
    await rejectTimesheet(rejectReason.value.trim())
    rejectOpen.value = false
    rejectReason.value = ''
    await loadAnomalies()
  } catch (err) {
    actionError.value = mapCraError(err)
  } finally {
    rejecting.value = false
  }
}

const onPrestationChange = (payload: PrestationInfoFields) => {
  prestation.client = payload.client
  prestation.mission = payload.mission
  prestation.clientId = payload.clientId
  prestation.missionId = payload.missionId
  prestation.description = payload.description
  prestation.technologies = payload.technologies
  prestation.lieu = payload.lieu
  prestation.responsableClient = payload.responsableClient
}

const persistPrestation = async () => {
  const local = (prestationFormRef.value?.local ?? prestation) as typeof prestation
  await apiFetch(`/api/cra/timesheets/${id.value}/commercial-info`, {
    method: 'PUT',
    body: {
      client: local.client,
      mission: local.mission,
      clientId: local.clientId || undefined,
      missionId: local.missionId || undefined,
      description: local.description,
      technologies: local.technologies,
      lieu: local.lieu,
      responsableClient: local.responsableClient
    }
  })
}

const savePrestation = async () => {
  savingPrestation.value = true
  prestationMsg.value = ''
  prestationError.value = false
  try {
    await persistPrestation()
    prestationMsg.value = t('cra.prestation_saved')
    await load()
  } catch {
    prestationMsg.value = t('cra.prestation_save_error')
    prestationError.value = true
  } finally {
    savingPrestation.value = false
  }
}

const ensurePrestationPersisted = async () => {
  if (!canEdit.value) return
  await persistPrestation()
}

const fetchPdfBlob = async () =>
  apiFetch<Blob>(`/api/cra/timesheets/${id.value}/pdf`, { method: 'POST', responseType: 'blob' })

const revokePdfPreviewUrl = () => {
  if (pdfPreviewUrl.value) {
    URL.revokeObjectURL(pdfPreviewUrl.value)
    pdfPreviewUrl.value = ''
  }
}

watch(pdfPreviewOpen, (open) => {
  if (!open) revokePdfPreviewUrl()
})

onUnmounted(revokePdfPreviewUrl)

const openPdfPreview = async () => {
  if (!canDownload.value) return
  pdfPreviewOpen.value = true
  pdfPreviewLoading.value = true
  pdfPreviewError.value = ''
  revokePdfPreviewUrl()
  try {
    await ensurePrestationPersisted()
    const blob = await fetchPdfBlob()
    pdfPreviewUrl.value = URL.createObjectURL(blob)
  } catch (err) {
    pdfPreviewError.value = mapCraError(err, t('cra.download_error'))
  } finally {
    pdfPreviewLoading.value = false
  }
}

const downloadPdf = async () => {
  if (!canDownload.value) return
  downloading.value = true
  downloadError.value = ''
  try {
    await ensurePrestationPersisted()
    const blob = await fetchPdfBlob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `cra-${timesheet.value?.month ?? id.value}.pdf`
    a.click()
    URL.revokeObjectURL(url)
  } catch (err) {
    downloadError.value = mapCraError(err, t('cra.download_error'))
  } finally {
    downloading.value = false
  }
}
</script>

<style scoped>
.cra-detail {
  display: grid;
  gap: var(--kore-space-lg);
}

@media (max-width: 768px) {
  .cra-detail :deep(.app-page-header__actions) {
    flex-wrap: wrap;
    gap: var(--kore-space-xs);
  }

  .cra-detail :deep(.app-page-header__actions .app-btn) {
    flex: 1 1 calc(50% - var(--kore-space-xs));
    min-width: 0;
  }
}

.meta {
  display: grid;
  gap: var(--kore-space-md);
  margin: 0;
}

.meta div {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--kore-space-md);
}

.meta dt {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.meta dd { margin: 0; font-weight: 600; }

.meta__empty {
  font-weight: 400;
  color: var(--kore-text-muted);
}

.meta__link {
  color: var(--kore-link);
  text-decoration: underline;
}

.cra-detail__body {
  display: grid;
  gap: var(--kore-space-lg);
  align-items: start;
}

.cra-detail__main {
  display: grid;
  gap: var(--kore-space-lg);
  min-width: 0;
}

.cra-detail__aside {
  min-width: 0;
}

@media (min-width: 900px) {
  .cra-detail__body {
    grid-template-columns: 1fr min(360px, 32%);
  }

  .cra-detail__aside {
    position: sticky;
    top: var(--kore-space-lg);
    order: 2;
  }

  .cra-detail__main {
    order: 1;
  }
}

@media (max-width: 768px) {
  .meta div {
    flex-direction: column;
    align-items: flex-start;
  }
}

.cra-detail__preview {
  margin-top: var(--kore-space-lg);
  padding-top: var(--kore-space-lg);
  border-top: 1px solid var(--kore-border);
}

.muted { color: var(--kore-text-muted); }

.flash--error { color: var(--kore-error); margin-top: var(--kore-space-md); }
.flash__link {
  margin-left: var(--kore-space-sm);
  color: var(--kore-link);
  text-decoration: underline;
}

.cra-detail__anomalies-title {
  margin: 0 0 var(--kore-space-sm);
  font-size: var(--kore-text-h3);
}

.cra-detail__anomalies-list {
  margin: 0;
  padding-left: 1.25rem;
  color: var(--kore-text);
  font-size: var(--kore-text-small);
}

.reject-form {
  display: grid;
  gap: var(--kore-space-md);
}

.reject-form__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  justify-content: flex-end;
}
</style>
