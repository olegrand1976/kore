<template>
  <div>
    <AppPageHeader :title="$t('ett.pointage_title')" :subtitle="$t('ett.pointage_subtitle')" />

    <AppCard padding="lg" class="pointage-card">
      <p class="pointage-card__clock">{{ nowLabel }}</p>
      <div class="pointage-card__actions">
        <AppButton variant="primary" :loading="clockingIn" :disabled="!!todayRecord?.clockIn" @click="clockIn">
          {{ $t('ett.clock_in') }}
        </AppButton>
        <AppButton variant="secondary" :loading="clockingOut" :disabled="!todayRecord?.clockIn || !!todayRecord?.clockOut" @click="clockOut">
          {{ $t('ett.clock_out') }}
        </AppButton>
      </div>
      <p v-if="message" class="pointage-card__msg">{{ message }}</p>
      <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>
    </AppCard>

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('fiche.loading') }}</p>
    </AppCard>

    <AppCard v-else padding="lg">
      <h3 class="pointage-section">{{ $t('ett.recent_records') }}</h3>
      <AppTable
        :columns="columns"
        :rows="rows"
        :empty-title="$t('ett.records_empty')"
      />
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const { t, locale } = useI18n()

type WorkRecord = {
  id: string
  workDate: string
  clockIn?: string | null
  clockOut?: string | null
  effectiveHours: number
  overtimeHours: number
}

const clockingIn = ref(false)
const clockingOut = ref(false)
const message = ref('')
const errorMsg = ref('')

const { data, pending, refresh } = await useAsyncData('ett-records', () =>
  $fetch<{ data?: WorkRecord[] }>('/api/ett/records')
)

const records = computed(() => data.value?.data ?? [])

const todayKey = new Date().toISOString().slice(0, 10)
const todayRecord = computed(() => records.value.find((r) => r.workDate?.slice(0, 10) === todayKey))

const nowLabel = computed(() =>
  new Date().toLocaleString(locale.value === 'en' ? 'en-US' : 'fr-FR', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    hour: '2-digit',
    minute: '2-digit'
  })
)

const columns = computed(() => [
  { key: 'workDate', label: t('ett.col_date') },
  { key: 'clockIn', label: t('ett.col_in') },
  { key: 'clockOut', label: t('ett.col_out') },
  { key: 'hours', label: t('ett.col_hours') },
  { key: 'overtime', label: t('ett.col_overtime') }
])

const formatTime = (raw?: string | null) => {
  if (!raw) return '—'
  return new Date(raw).toLocaleTimeString(locale.value === 'en' ? 'en-US' : 'fr-FR', {
    hour: '2-digit',
    minute: '2-digit'
  })
}

const rows = computed(() =>
  records.value.map((record) => ({
    id: record.id,
    workDate: record.workDate?.slice(0, 10) ?? '—',
    clockIn: formatTime(record.clockIn),
    clockOut: formatTime(record.clockOut),
    hours: record.effectiveHours?.toFixed(2) ?? '0',
    overtime: record.overtimeHours?.toFixed(2) ?? '0'
  }))
)

async function clockIn() {
  errorMsg.value = ''
  message.value = ''
  clockingIn.value = true
  try {
    await $fetch('/api/ett/clock-in', { method: 'POST', body: {} })
    message.value = t('ett.clock_in_ok')
    await refresh()
  } catch {
    errorMsg.value = t('ett.clock_error')
  } finally {
    clockingIn.value = false
  }
}

async function clockOut() {
  errorMsg.value = ''
  message.value = ''
  clockingOut.value = true
  try {
    await $fetch('/api/ett/clock-out', { method: 'POST', body: {} })
    message.value = t('ett.clock_out_ok')
    await refresh()
  } catch {
    errorMsg.value = t('ett.clock_error')
  } finally {
    clockingOut.value = false
  }
}
</script>

<style scoped>
.pointage-card {
  margin-bottom: var(--kore-space-lg);
}

.pointage-card__clock {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-h3);
  font-weight: 600;
}

.pointage-card__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
}

.pointage-card__msg {
  margin: var(--kore-space-md) 0 0;
  color: var(--kore-success);
}

.pointage-section {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-body);
}

.muted {
  margin: 0;
  color: var(--kore-text-muted);
}

.flash--error {
  color: var(--kore-error);
}

@media (max-width: 768px) {
  .pointage-card__actions :deep(.app-button) {
    flex: 1 1 100%;
  }
}
</style>
