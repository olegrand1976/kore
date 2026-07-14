<template>
  <div>
    <AppPageHeader :title="pageTitle">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/cra')">
          <AppIcon name="arrow_back" /> {{ $t('cra.back') }}
        </AppButton>
        <AppButton
          v-if="isAdmin"
          variant="secondary"
          size="sm"
          :disabled="!canEdit || saving"
          @click="validateFinal"
        >
          {{ $t('cra.validate_final') }}
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
          variant="primary"
          size="sm"
          :disabled="downloading || !canDownload"
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
      <AppCard padding="lg" class="cra-detail__meta">
        <dl class="meta">
          <div><dt>{{ $t('cra.period') }}</dt><dd>{{ formatMonth(timesheet.month) }}</dd></div>
          <div>
            <dt>{{ $t('cra.col_status') }}</dt>
            <dd><AppBadge :variant="statusVariant(timesheet.status)">{{ statusLabel(timesheet.status) }}</AppBadge></dd>
          </div>
        </dl>
        <CraMonthlyPreview
          class="cra-detail__preview"
          :total-minutes="monthStats.totalMinutes"
          :capacity-minutes="monthStats.capacityMinutes"
          :weeks-submitted="monthStats.weeksSubmitted"
          :weeks-total="monthStats.weeksTotal"
          :prefill-ratio="monthStats.prefillRatio"
          :progress="monthStats.progress"
        />
      </AppCard>

      <TimesheetGrid
        :weeks="selectedWeeks"
        :month="timesheet.month"
        :week-start-day="weekStartDay"
        :day-capacity-minutes="dayCapacityMinutes"
        :week-submit-policy="weekSubmitPolicy"
        :can-edit="canEdit"
        :saving="saving"
        :missions="missions"
        @save="onSaveWeek"
        @submit="onSubmitWeek"
      />

      <CommercialInfoForm
        :client="commercial.client"
        :mission="commercial.mission"
        :description="commercial.description"
        :technologies="commercial.technologies"
        :lieu="commercial.lieu"
        :responsable-client="commercial.responsableClient"
        :disabled="!canEdit"
        :saving="savingCommercial"
        :message="commercialMsg"
        :is-error="commercialError"
        @submit="saveCommercial"
      />
    </div>

    <p v-if="prefillMsg" class="flash flash--info" role="status">{{ prefillMsg }}</p>
    <p v-if="downloadError" class="flash flash--error" role="alert">{{ downloadError }}</p>
  </div>
</template>

<script setup lang="ts">
import type { CraLine } from '~/stores/cra'
import { weekNumberForDay } from '~/composables/useWeekCalendar'
import { useCraMonthStats } from '~/composables/useCraMonthStats'

definePageMeta({ layout: 'default' })

const route = useRoute()
const { t, locale } = useI18n()
const { statusLabel, statusVariant } = useCraStatus()
const { isAdmin } = useAuth()
const id = computed(() => String(route.params.id))

const { timesheet, loading, error, canEdit, selectedWeeks, saving, load, saveWeek, submitWeek, validateFinal } = useCra(id)

const weekStartDay = ref(1)
const dayCapacityMinutes = ref(480)
const weekSubmitPolicy = ref<'block' | 'warn' | 'none'>('warn')
const missions = ref<Array<{ id: string; clientName?: string; clientId?: string }>>([])

const loadOrgSettings = async () => {
  try {
    const res = await $fetch<{
      data?: {
        weekStartDay?: number
        dayCapacityMinutes?: number
        weekSubmitPolicy?: string
      }
      weekStartDay?: number
      dayCapacityMinutes?: number
      weekSubmitPolicy?: string
    }>('/api/org/users/me/calendar-settings')
    const data = res.data ?? res
    const day = data.weekStartDay
    if (day != null && day >= 0 && day <= 6) {
      weekStartDay.value = day
    }
    if (data.dayCapacityMinutes != null && data.dayCapacityMinutes > 0) {
      dayCapacityMinutes.value = data.dayCapacityMinutes
    }
    const policy = data.weekSubmitPolicy
    if (policy === 'block' || policy === 'warn' || policy === 'none') {
      weekSubmitPolicy.value = policy
    }
  } catch {
    weekStartDay.value = 1
    dayCapacityMinutes.value = 480
    weekSubmitPolicy.value = 'warn'
  }
}

const loadMissions = async () => {
  try {
    const res = await $fetch<{ data: Array<Record<string, unknown>> }>('/api/ssii/missions')
    missions.value = (res.data ?? []).map((item) => ({
      id: String(item.id ?? item.ID ?? ''),
      clientName: String(item.clientName ?? item.ClientName ?? ''),
      clientId: String(item.clientId ?? item.ClientID ?? '')
    })).filter((m) => m.id)
  } catch {
    missions.value = []
  }
}

await Promise.all([load(id.value), loadOrgSettings(), loadMissions()])

const monthRef = computed(() => timesheet.value?.month ?? '')
const weekStartDayRef = computed(() => weekStartDay.value)
const weeksRef = computed(() => selectedWeeks.value)
const monthStats = useCraMonthStats(weeksRef, monthRef, weekStartDayRef, dayCapacityMinutes)

const commercial = reactive({
  client: '',
  mission: '',
  description: '',
  technologies: [] as string[],
  lieu: '',
  responsableClient: ''
})
const savingCommercial = ref(false)
const commercialMsg = ref('')
const commercialError = ref(false)
const downloading = ref(false)
const downloadError = ref('')
const prefillLoading = ref(false)
const prefillMsg = ref('')
const { suggestCraPrefill, extractFetchError: aiError } = useAi()

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
    const res = await $fetch<{ data?: { added?: number } }>(`/api/cra/timesheets/${id.value}/prefill-holidays`, {
      method: 'POST'
    })
    await load(id.value)
    prefillMsg.value = t('cra.prefill_holidays_result', { n: res?.data?.added ?? 0 })
  } catch (err) {
    downloadError.value = aiError(err)
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
    downloadError.value = aiError(err)
  } finally {
    prefillLoading.value = false
  }
}

watch(timesheet, (ts) => {
  if (!ts?.commercialInfo) return
  commercial.client = ts.commercialInfo.client ?? ''
  commercial.mission = ts.commercialInfo.mission ?? ''
  commercial.description = ts.commercialInfo.description ?? ''
  commercial.technologies = [...(ts.commercialInfo.technologies ?? [])]
  commercial.lieu = ts.commercialInfo.lieu ?? ''
  commercial.responsableClient = ts.commercialInfo.responsableClient ?? ''
}, { immediate: true })

const canDownload = computed(() => Boolean(commercial.client.trim() && commercial.mission.trim()))

const pageTitle = computed(() => {
  if (!timesheet.value?.month) return t('cra.title')
  return `${t('cra.title')} — ${formatMonth(timesheet.value.month)}`
})

const formatMonth = (raw: string) => {
  const [y, m] = raw.split('-').map(Number)
  return new Date(y, m - 1, 1).toLocaleDateString(locale.value === 'en' ? 'en-US' : 'fr-FR', {
    month: 'long', year: 'numeric'
  })
}

const onSaveWeek = async (weekNumber: number, lines: Parameters<typeof saveWeek>[1]) => {
  await saveWeek(weekNumber, lines)
}

const onSubmitWeek = async (weekNumber: number) => {
  await submitWeek(weekNumber)
}

const saveCommercial = async () => {
  savingCommercial.value = true
  commercialMsg.value = ''
  commercialError.value = false
  try {
    await $fetch(`/api/cra/timesheets/${id.value}/commercial-info`, {
      method: 'PUT',
      body: {
        client: commercial.client,
        mission: commercial.mission,
        description: commercial.description,
        technologies: commercial.technologies,
        lieu: commercial.lieu,
        responsableClient: commercial.responsableClient
      }
    })
    commercialMsg.value = t('cra.commercial_saved')
    await load()
  } catch {
    commercialMsg.value = t('cra.download_error')
    commercialError.value = true
  } finally {
    savingCommercial.value = false
  }
}

const downloadPdf = async () => {
  if (!canDownload.value) return
  downloading.value = true
  downloadError.value = ''
  try {
    const blob = await $fetch<Blob>(`/api/cra/timesheets/${id.value}/pdf`, { method: 'POST', responseType: 'blob' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `cra-${timesheet.value?.month ?? id.value}.pdf`
    a.click()
    URL.revokeObjectURL(url)
  } catch {
    downloadError.value = t('cra.download_error')
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

.cra-detail__preview {
  margin-top: var(--kore-space-lg);
  padding-top: var(--kore-space-lg);
  border-top: 1px solid var(--kore-border);
}

.muted { color: var(--kore-text-muted); }

.flash--error { color: var(--kore-error); margin-top: var(--kore-space-md); }
</style>
