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
          :disabled="!canEdit || saving || validatingFinal || forceValidating"
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
          :disabled="downloading"
          :aria-describedby="hasPrestationContext ? undefined : 'cra-download-hint'"
          @click="openPdfPreview"
        >
          {{ $t('cra.preview_pdf') }}
        </AppButton>
        <AppButton
          variant="primary"
          size="sm"
          :disabled="downloading"
          :aria-describedby="hasPrestationContext ? undefined : 'cra-download-hint'"
          @click="downloadPdf"
        >
          <AppIcon name="download" /> {{ $t('cra.download') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <!--
      Les retours d'action restent sous l'en-tête, au contact des boutons qui les
      déclenchent : plus bas, ils tombaient sous la grille du mois et un échec de
      validation définitive passait pour un bouton sans effet.
    -->
    <div v-if="openNotice || validateMsg || prefillMsg || downloadError || actionError" class="cra-detail__flashes">
      <p v-if="openNotice" class="flash flash--info" role="status">{{ openNotice }}</p>
      <p v-if="validateMsg" class="flash" role="status">
        {{ validateMsg }}
        <NuxtLink v-if="invoiceLink" :to="invoiceLink" class="flash__link">{{ $t('cra.invoice_created_link') }}</NuxtLink>
      </p>
      <p v-if="prefillMsg" class="flash flash--info" role="status">{{ prefillMsg }}</p>
      <p v-if="downloadError" class="flash flash--error" role="alert">{{ downloadError }}</p>
      <p v-if="actionError" class="flash flash--error" role="alert">{{ actionError }}</p>
    </div>

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
          <div>
            <dt>{{ $t('cra.col_user') }}</dt>
            <dd>
              <NuxtLink
                v-if="timesheet.userId"
                :to="`/collaborateurs/${timesheet.userId}`"
                class="meta__link"
              >
                {{ ownerLabel || timesheet.userId }}
              </NuxtLink>
              <span v-else>{{ ownerLabel || $t('common.none') }}</span>
            </dd>
          </div>
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
        <div v-if="canLinkMission" class="cra-detail__mission-link">
          <label for="cra-mission-link">{{ $t('cra.mission_link_label') }}</label>
          <select
            id="cra-mission-link"
            v-model="missionLinkId"
            :disabled="linkingMission"
            @change="onMissionLinkChange"
          >
            <option v-if="canEdit" value="">{{ $t('cra.mission_link_none') }}</option>
            <option v-for="mission in missions" :key="mission.id" :value="mission.id">
              {{ formatMissionOptionLabel(mission) }}
            </option>
          </select>
          <p class="cra-detail__mission-hint">{{ $t('cra.mission_link_hint') }}</p>
          <p v-if="missionLinkError" class="flash flash--error" role="alert">{{ missionLinkError }}</p>
        </div>
        <p v-if="!hasPrestationContext" id="cra-download-hint" class="cra-detail__download-hint">
          {{ $t('cra.download_hint_mission') }}
        </p>
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

      <div class="cra-detail__body cra-detail__body--single">
        <div class="cra-detail__main">
          <TimesheetGrid
            v-model:active-week="gridActiveWeek"
            :weeks="selectedWeeks"
            :month="timesheet.month"
            :week-start-day="weekStartDay"
            :day-capacity-minutes="dayCapacityMinutes"
            :week-submit-policy="weekSubmitPolicy"
            :can-edit="canEdit"
            :saving="saving"
            :planned-week-minutes="missionPlannedWeekMinutes"
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

    <CraForceValidateModal
      v-model:open="forceValidateOpen"
      :validating="forceValidating"
      @confirm="confirmForceValidate"
    />
  </div>
</template>

<script setup lang="ts">
import type { CraLine } from '~/stores/cra'
import { weekNumberForDay } from '~/composables/useWeekCalendar'
import { useCraMonthStats } from '~/composables/useCraMonthStats'
import { useCraWorkRefs } from '~/composables/useCraWorkRefs'
import {
  prestationInfoComplete,
  canLinkCraMission,
  unwrapMissionPayload,
  missionPrestationPatch,
  formatMissionOptionLabel,
  mapMissionListItem,
  missionTitleFromPayload
} from '~/utils/craPrestation'
import { normalizeAnomalyMessages } from '~/utils/craAnomalies'
import { timesheetHasLoggedTime } from '~/utils/craLoggedTime'
import { formatUserDisplayName } from '~/composables/useUserDisplay'

definePageMeta({ layout: 'default' })

const { apiFetch } = useApiFetch()
const route = useRoute()
const { t, locale } = useI18n()
const { statusLabel, statusVariant } = useCraStatus()
const { can, canValidateCra } = usePermissions()
const { mapCraError, mapInvoiceDraftMessage: mapInvoiceDraft, isCommercialInfoRequiredError } = useCraError()
const id = computed(() => String(route.params.id))

const { timesheet, loading, error, canEdit, selectedWeeks, saving, load, saveWeek, submitWeek, validateFinal, rejectTimesheet } = useCra(id)
const { user, fetchSession } = useAuth()
const { options: workRefOptions, load: loadWorkRefs, labelFor: workRefLabelFor } = useCraWorkRefs()

/** Locked week tab for TimesheetGrid (null = auto current week). Survives save remounts. */
const gridActiveWeek = ref<number | null>(null)
watch(id, () => {
  gridActiveWeek.value = null
})
watch(() => timesheet.value?.month, (month, prev) => {
  if (prev != null && month != null && month !== prev) {
    gridActiveWeek.value = null
  }
})

const weekStartDay = ref(1)
const dayCapacityMinutes = ref(480)
const weekSubmitPolicy = ref<'block' | 'warn' | 'none'>('warn')
const taskTypesEnabled = ref<string[]>(['manual', 'interne', 'formation', 'mission'])
const missionPlannedWeekMinutes = ref<number | null>(null)
const missions = ref<Array<{ id: string; clientName?: string; clientId?: string; title?: string; label?: string }>>([])
const missionLinkId = ref('')
const linkingMission = ref(false)
const missionLinkError = ref('')
const ownerLabel = ref('')
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
    missions.value = (res.data ?? [])
      .map((item) => mapMissionListItem(item))
      .filter((m) => m.id)
  } catch {
    missions.value = []
  }
}

const loadMissionPlannedWeekMinutes = async (missionId: string) => {
  if (!missionId) {
    missionPlannedWeekMinutes.value = null
    return
  }
  try {
    const res = await apiFetch<{ data?: { plannedWeekMinutes?: number | null }; plannedWeekMinutes?: number | null }>(
      `/api/ssii/missions/${missionId}`
    )
    const data = res.data ?? res
    const planned = data.plannedWeekMinutes
    missionPlannedWeekMinutes.value =
      planned != null && Number.isFinite(Number(planned)) && Number(planned) > 0
        ? Number(planned)
        : null
  } catch {
    missionPlannedWeekMinutes.value = null
  }
}

const downloading = ref(false)
const downloadError = ref('')
const prefillLoading = ref(false)
const prefillMsg = ref('')
const actionError = ref('')
const validateMsg = ref('')
const openNotice = ref('')
const invoiceLink = ref('')
const rejectOpen = ref(false)
const rejectReason = ref('')
const rejecting = ref(false)
const rejectTitleId = 'cra-reject-title'
const rejectReasonId = 'cra-reject-reason'
const forceValidateOpen = ref(false)
const forceValidating = ref(false)
const validatingFinal = ref(false)
const anomalies = ref<string[]>([])
const anomaliesLoading = ref(false)
const { suggestCraPrefill, fetchCraAnomalies } = useAi()

const loadAnomalies = async () => {
  anomaliesLoading.value = true
  try {
    const res = await fetchCraAnomalies(id.value)
    anomalies.value = normalizeAnomalyMessages(res)
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

const workRefsOwnerId = computed(() => timesheet.value?.userId || '')
watch(
  workRefsOwnerId,
  async (ownerId) => {
    if (!ownerId) return
    await loadWorkRefs(ownerId)
  },
  { immediate: true }
)

watch(
  () => timesheet.value?.userId,
  async (uid) => {
    ownerLabel.value = ''
    if (!uid) return
    try {
      const res = await apiFetch<{
        data?: { prenom?: string; nom?: string; login?: string; Prenom?: string; Nom?: string; Login?: string }
      }>(`/api/org/users/${uid}`)
      const data = res.data ?? (res as unknown as Record<string, string>)
      ownerLabel.value = formatUserDisplayName(
        data.prenom ?? data.Prenom,
        data.nom ?? data.Nom,
        data.login ?? data.Login
      )
    } catch {
      ownerLabel.value = uid.slice(0, 8)
    }
  },
  { immediate: true }
)
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
  if (!ts?.commercialInfo) {
    prestation.client = ''
    prestation.mission = ''
    prestation.clientId = ''
    prestation.missionId = ''
    prestation.description = ''
    prestation.technologies = []
    prestation.lieu = ''
    prestation.responsableClient = ''
    missionLinkId.value = ''
    return
  }
  prestation.client = ts.commercialInfo.client ?? ''
  prestation.mission = ts.commercialInfo.mission ?? ''
  prestation.clientId = ts.commercialInfo.clientId ?? ''
  prestation.missionId = ts.commercialInfo.missionId ?? ''
  prestation.description = ts.commercialInfo.description ?? ''
  prestation.technologies = [...(ts.commercialInfo.technologies ?? [])]
  prestation.lieu = ts.commercialInfo.lieu ?? ''
  prestation.responsableClient = ts.commercialInfo.responsableClient ?? ''
  missionLinkId.value = prestation.missionId
}, { immediate: true })

watch(
  () => prestation.missionId,
  (missionId) => {
    void loadMissionPlannedWeekMinutes(missionId || '')
  },
  { immediate: true }
)

const hasPrestationContext = computed(() => prestationInfoComplete(prestation.client, prestation.mission))
/** Full edits while editable; fill-once after Définitif when client/mission still missing. */
const canLinkMission = computed(() =>
  canLinkCraMission({
    canWrite: can('cra', 'E'),
    canEditTimesheet: canEdit.value,
    prestationComplete: hasPrestationContext.value
  })
)

const persistMissionLink = async (missionId: string) => {
  if (!missionId) {
    await apiFetch(`/api/cra/timesheets/${id.value}/commercial-info`, {
      method: 'PUT',
      body: {
        client: '',
        mission: '',
        description: '',
        technologies: [],
        lieu: '',
        responsableClient: ''
      }
    })
    return
  }
  const res = await apiFetch(`/api/ssii/missions/${missionId}`)
  const raw = unwrapMissionPayload(res)
  const patch = missionPrestationPatch(raw)
  const listed = missions.value.find((m) => m.id === missionId)
  const missionLabel =
    missionTitleFromPayload(raw) ||
    listed?.title?.trim() ||
    listed?.label?.trim() ||
    patch.client ||
    missionId
  await apiFetch(`/api/cra/timesheets/${id.value}/commercial-info`, {
    method: 'PUT',
    body: {
      client: patch.client || listed?.clientName || '',
      mission: missionLabel,
      clientId: patch.clientId || undefined,
      missionId,
      description: '',
      technologies: patch.technologies,
      lieu: '',
      responsableClient: patch.responsableClient
    }
  })
}

const onMissionLinkChange = async () => {
  if (!canLinkMission.value) return
  const previousMissionId = prestation.missionId
  const nextMissionId = missionLinkId.value
  linkingMission.value = true
  missionLinkError.value = ''
  try {
    await persistMissionLink(nextMissionId)
    await load(id.value)
  } catch (err) {
    missionLinkId.value = previousMissionId
    missionLinkError.value = mapCraError(err, t('cra.mission_link_error'))
  } finally {
    linkingMission.value = false
  }
}

const pageTitle = computed(() => {
  if (!timesheet.value?.month) return t('cra.title')
  const period = formatMonth(timesheet.value.month)
  if (ownerLabel.value) {
    return t('cra.detail_title_owner', { period, user: ownerLabel.value })
  }
  return t('cra.detail_title', { period })
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

const pendingOpenNotice = (() => {
  const raw = route.query.notice
  const value = Array.isArray(raw) ? raw[0] : raw
  if (value === 'created' || value === 'existing') return value
  return null
})()

if (pendingOpenNotice) {
  const nextQuery = { ...route.query }
  delete nextQuery.notice
  void navigateTo({ path: route.path, query: nextQuery }, { replace: true })
}

watch(
  () => timesheet.value?.month,
  (month) => {
    if (!pendingOpenNotice || !month || openNotice.value) return
    const period = formatMonth(String(month))
    switch (pendingOpenNotice) {
      case 'created':
        openNotice.value = t('cra.notice_created', { period })
        break
      case 'existing':
        openNotice.value = t('cra.notice_existing', { period })
        break
      default: {
        const _exhaustive: never = pendingOpenNotice
        return _exhaustive
      }
    }
  },
  { immediate: true }
)

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

const hasLoggedTime = computed(() => timesheetHasLoggedTime(timesheet.value?.weeks))

const applyValidateSuccess = async (draft: Awaited<ReturnType<typeof validateFinal>>) => {
  validateMsg.value = mapInvoiceDraft(draft)
  const invoiceId = (draft as { invoiceId?: string } | undefined)?.invoiceId
  if (draft?.status === 'created' && invoiceId) {
    invoiceLink.value = `/facturation/${invoiceId}`
  }
  await loadAnomalies()
}

const onValidateFinal = async () => {
  if (validatingFinal.value || forceValidating.value) return
  validatingFinal.value = true
  actionError.value = ''
  validateMsg.value = ''
  invoiceLink.value = ''
  try {
    const draft = await validateFinal()
    await applyValidateSuccess(draft)
  } catch (err) {
    // Infos prestation manquantes + temps saisi : modal de forçage sans banner d'erreur concurrente.
    if (isCommercialInfoRequiredError(err) && hasLoggedTime.value) {
      forceValidateOpen.value = true
      return
    }
    actionError.value = mapCraError(err)
  } finally {
    validatingFinal.value = false
  }
}

const confirmForceValidate = async () => {
  if (forceValidating.value || validatingFinal.value) return
  forceValidating.value = true
  actionError.value = ''
  validateMsg.value = ''
  invoiceLink.value = ''
  try {
    const draft = await validateFinal({ force: true })
    forceValidateOpen.value = false
    await applyValidateSuccess(draft)
  } catch (err) {
    // Fermer la modal pour n'afficher qu'un seul signal d'erreur (flash).
    forceValidateOpen.value = false
    actionError.value = mapCraError(err)
  } finally {
    forceValidating.value = false
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

const fetchPdfBlob = async () => {
  const blob = await apiFetch<Blob>(`/api/cra/timesheets/${id.value}/pdf`, { method: 'POST', responseType: 'blob' })
  // Un relais binaire cassé renvoie 200 avec un corps vide : sans ce garde-fou
  // le navigateur télécharge un .pdf illisible au lieu d'afficher une erreur.
  if (!blob || blob.size === 0) throw new Error('empty pdf payload')
  return blob
}

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
  pdfPreviewOpen.value = true
  pdfPreviewLoading.value = true
  pdfPreviewError.value = ''
  revokePdfPreviewUrl()
  try {
    const blob = await fetchPdfBlob()
    pdfPreviewUrl.value = URL.createObjectURL(blob)
  } catch (err) {
    pdfPreviewError.value = mapCraError(err, t('cra.download_error'))
  } finally {
    pdfPreviewLoading.value = false
  }
}

const downloadPdf = async () => {
  downloading.value = true
  downloadError.value = ''
  try {
    const blob = await fetchPdfBlob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `cra-${timesheet.value?.month ?? id.value}.pdf`
    // Firefox et Safari ignorent un lien détaché du document et annulent le
    // téléchargement si l'URL est révoquée dans la foulée du clic.
    a.style.display = 'none'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    setTimeout(() => URL.revokeObjectURL(url), 0)
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

.cra-detail__body--single {
  grid-template-columns: 1fr;
}

.cra-detail__main {
  display: grid;
  gap: var(--kore-space-lg);
  min-width: 0;
}

.cra-detail__mission-link {
  display: grid;
  gap: var(--kore-space-xs);
  margin-top: var(--kore-space-md);
}

.cra-detail__mission-link label {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.cra-detail__mission-link select {
  width: 100%;
  max-width: var(--kore-form-max);
  min-height: 2.5rem;
  padding: var(--kore-space-xs) var(--kore-space-sm);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-sm);
  background: var(--kore-surface);
  color: var(--kore-text);
}

.cra-detail__mission-hint,
.cra-detail__download-hint {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

@media (max-width: 768px) {
  .meta div {
    flex-direction: column;
    align-items: flex-start;
  }

  .cra-detail__mission-link select {
    max-width: none;
  }
}

.cra-detail__preview {
  margin-top: var(--kore-space-lg);
  padding-top: var(--kore-space-lg);
  border-top: 1px solid var(--kore-border);
}

.muted { color: var(--kore-text-muted); }

.cra-detail__flashes {
  display: grid;
  gap: var(--kore-space-xs);
  margin-bottom: var(--kore-space-md);
}

.flash--error { color: var(--kore-error); margin-top: var(--kore-space-md); }
.flash--info { color: var(--kore-brand-blue); margin-top: var(--kore-space-md); }
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
