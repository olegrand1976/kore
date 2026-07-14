<template>
  <div>
    <AppPageHeader :title="$t('ett.reconciliation_title')" />

    <AppCard padding="lg" class="ett-filters">
      <label for="ett-month">{{ $t('prestations.month') }}</label>
      <input id="ett-month" v-model="month" type="month" @change="refresh">
    </AppCard>

    <AppCard v-if="pending" padding="lg">
      <CraSkeleton />
    </AppCard>

    <AppCard v-else-if="errorMsg" padding="lg">
      <p class="flash flash--error">{{ errorMsg }}</p>
    </AppCard>

    <AppCard v-else-if="report" padding="lg" class="ett-report">
      <AppBadge :variant="report.alert ? 'warning' : 'success'">
        {{ report.alert ? $t('ett.alert') : $t('ett.ok') }}
      </AppBadge>
      <dl class="ett-report__grid">
        <div>
          <dt>{{ $t('ett.cra_hours') }}</dt>
          <dd>{{ formatHours(report.craHours) }}</dd>
        </div>
        <div>
          <dt>{{ $t('ett.ett_hours') }}</dt>
          <dd>{{ formatHours(report.ettHours) }}</dd>
        </div>
        <div>
          <dt>{{ $t('ett.delta_hours') }}</dt>
          <dd>{{ formatHours(report.deltaHours) }}</dd>
        </div>
      </dl>
      <p v-if="report.alertMessage" class="ett-report__message">{{ report.alertMessage }}</p>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const { t } = useI18n()

type ReconciliationReport = {
  craHours?: number
  ettHours?: number
  deltaHours?: number
  alert?: boolean
  alertMessage?: string
}

const month = ref(new Date().toISOString().slice(0, 7))
const errorMsg = ref('')

const { data, pending, refresh, error } = await useFetch<{ data?: ReconciliationReport }>(
  () => `/api/ett/reconciliation?month=${month.value}`
)

watch(error, (err) => {
  errorMsg.value = err ? t('ett.load_error') : ''
})

const report = computed(() => data.value?.data ?? null)

const formatHours = (value?: number) => {
  if (value == null) return '—'
  return t('cra.hours_value', { n: Math.round(value * 10) / 10 })
}
</script>

<style scoped>
.ett-filters {
  display: flex;
  align-items: center;
  gap: var(--kore-space-md);
  margin-bottom: var(--kore-space-lg);
}

.ett-filters label {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.ett-report__grid {
  display: grid;
  gap: var(--kore-space-md);
  margin-top: var(--kore-space-lg);
}

@media (min-width: 769px) {
  .ett-report__grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

.ett-report__grid dt {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.ett-report__grid dd {
  margin: 0;
  font-size: var(--kore-text-h3);
  font-weight: 600;
}

.ett-report__message {
  margin-top: var(--kore-space-md);
  color: var(--kore-text-muted);
}

.flash--error {
  color: var(--kore-error);
}
</style>
