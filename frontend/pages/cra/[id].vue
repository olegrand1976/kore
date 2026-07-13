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
      <p class="muted">{{ $t('cra.loading') }}</p>
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
      </AppCard>

      <TimesheetGrid
        :weeks="selectedWeeks"
        :month="timesheet.month"
        :can-edit="canEdit"
        :saving="saving"
        @save="onSaveWeek"
        @submit="onSubmitWeek"
      />

      <CommercialInfoForm
        :client="commercial.client"
        :mission="commercial.mission"
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
definePageMeta({ layout: 'default' })

const route = useRoute()
const { t, locale } = useI18n()
const { statusLabel, statusVariant } = useCraStatus()
const { isAdmin } = useAuth()
const id = computed(() => String(route.params.id))

const { timesheet, loading, error, canEdit, selectedWeeks, saving, load, saveWeek, submitWeek, validateFinal } = useCra(id)

await load(id.value)

const commercial = reactive({ client: '', mission: '' })
const savingCommercial = ref(false)
const commercialMsg = ref('')
const commercialError = ref(false)
const downloading = ref(false)
const downloadError = ref('')
const prefillLoading = ref(false)
const prefillMsg = ref('')
const { suggestCraPrefill, extractFetchError: aiError } = useAi()

const loadPrefillSuggest = async () => {
  prefillLoading.value = true
  prefillMsg.value = ''
  try {
    const res = await suggestCraPrefill(id.value)
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
      body: { client: commercial.client, mission: commercial.mission }
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
    a.download = `cra-${timesheet.value?.month ?? id.value}.html`
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

.muted { color: var(--kore-text-muted); }

.flash--error { color: var(--kore-error); margin-top: var(--kore-space-md); }
</style>
