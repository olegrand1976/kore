<template>
  <div>
    <AppPageHeader :title="pageTitle">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/cra')">
          <AppIcon name="arrow_back" /> {{ $t('cra.back') }}
        </AppButton>
        <AppButton
          variant="primary"
          size="sm"
          :disabled="downloading || !canDownload"
          :title="canDownload ? undefined : $t('cra.download_hint')"
          @click="downloadPdf"
        >
          <AppIcon name="download" /> {{ $t('cra.download') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('cra.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="error" padding="lg">
      <AppEmptyState icon="error" :title="$t('cra.not_found')" />
    </AppCard>

    <div v-else-if="timesheet" class="cra-detail">
      <AppCard padding="lg" class="cra-detail__meta">
        <dl class="meta">
          <div><dt>{{ $t('cra.period') }}</dt><dd>{{ formatMonth(String(timesheet.month)) }}</dd></div>
          <div>
            <dt>{{ $t('cra.col_status') }}</dt>
            <dd><AppBadge :variant="statusVariant(String(timesheet.status))">{{ statusLabel(String(timesheet.status)) }}</AppBadge></dd>
          </div>
          <div><dt>{{ $t('cra.weeks') }}</dt><dd>{{ timesheet.weeks?.length ?? 0 }}</dd></div>
        </dl>
      </AppCard>

      <AppCard padding="lg" class="cra-detail__commercial">
        <h3 class="section-title">{{ $t('cra.commercial_title') }}</h3>
        <p class="section-hint">{{ $t('cra.commercial_hint') }}</p>
        <form class="commercial-form" @submit.prevent="saveCommercial">
          <AppInput id="client" v-model="commercial.client" :label="$t('cra.client')" />
          <AppInput id="mission" v-model="commercial.mission" :label="$t('cra.mission')" />
          <AppButton variant="primary" size="sm" type="submit" :disabled="savingCommercial">
            {{ $t('cra.save_commercial') }}
          </AppButton>
        </form>
        <p v-if="commercialMsg" class="flash" :class="{ 'flash--error': commercialError }" role="status">{{ commercialMsg }}</p>
        <p v-if="!canDownload" class="hint">{{ $t('cra.download_hint') }}</p>
      </AppCard>
    </div>

    <p v-if="downloadError" class="flash flash--error" role="alert">{{ downloadError }}</p>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const route = useRoute()
const { t, locale } = useI18n()
const { statusLabel, statusVariant } = useCraStatus()
const id = computed(() => String(route.params.id))

const { data, pending, error, refresh } = await useFetch(() => `/api/cra/timesheets/${id.value}`)
const timesheet = computed(() => (data.value as any)?.data ?? data.value)

const commercial = reactive({ client: '', mission: '' })
const savingCommercial = ref(false)
const commercialMsg = ref('')
const commercialError = ref(false)
const downloading = ref(false)
const downloadError = ref('')

watch(timesheet, (ts) => {
  if (!ts?.commercialInfo) return
  commercial.client = ts.commercialInfo.client ?? ''
  commercial.mission = ts.commercialInfo.mission ?? ''
}, { immediate: true })

const canDownload = computed(() => Boolean(commercial.client.trim() && commercial.mission.trim()))

const pageTitle = computed(() => {
  if (!timesheet.value?.month) return t('cra.title')
  return `${t('cra.title')} — ${formatMonth(String(timesheet.value.month))}`
})

const formatMonth = (raw: string) => {
  const [y, m] = raw.split('-').map(Number)
  return new Date(y, m - 1, 1).toLocaleDateString(locale.value === 'en' ? 'en-US' : 'fr-FR', {
    month: 'long', year: 'numeric'
  })
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
    await refresh()
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

@media (max-width: 480px) {
  .meta div {
    flex-direction: column;
    align-items: flex-start;
  }

  .commercial-form {
    max-width: none;
  }
}

.meta dt {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.meta dd { margin: 0; font-weight: 600; }

.section-title {
  margin: 0 0 var(--kore-space-xs);
  font-size: var(--kore-text-h3);
}

.section-hint {
  margin: 0 0 var(--kore-space-lg);
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.commercial-form {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
  max-width: 420px;
}

.hint {
  margin: var(--kore-space-md) 0 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.muted { color: var(--kore-text-muted); }

.flash {
  margin-top: var(--kore-space-md);
  font-size: var(--kore-text-small);
  color: var(--kore-success);
}

.flash--error { color: var(--kore-error); }
</style>
