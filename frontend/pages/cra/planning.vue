<template>
  <div>
    <AppPageHeader :title="$t('cra.planning_title')" :subtitle="$t('cra.planning_subtitle')">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/cra/gantt')">
          {{ $t('cra.gantt_link') }}
        </AppButton>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/cra')">
          <AppIcon name="arrow_back" /> {{ $t('cra.back') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard padding="lg">
      <p v-if="pending" class="muted">{{ $t('cra.loading') }}</p>
      <p v-else-if="error" class="flash flash--error" role="alert">{{ error }}</p>
      <PlanningBoard
        v-else
        :rows="rows"
        :days="days"
        :user-header="$t('cra.planning_col_user')"
        :empty-title="$t('cra.planning_empty')"
      />
    </AppCard>
  </div>
</template>

<script setup lang="ts">
import type { PlanningRow } from '~/composables/useReporting'

definePageMeta({ layout: 'default' })

const { fetchPlanning, rollingWindow60 } = useReporting()
const { t } = useI18n()

const { data, pending, error: fetchError } = await useAsyncData('cra-planning', () => fetchPlanning({ window: '60' }))

const rows = computed(() => data.value ?? [])

const days = computed(() => {
  const period = rollingWindow60()
  const start = new Date(`${period.start}T00:00:00Z`)
  const end = new Date(`${period.end}T00:00:00Z`)
  const out: string[] = []
  for (let d = new Date(start); d <= end; d.setUTCDate(d.getUTCDate() + 1)) {
    out.push(d.toISOString().slice(0, 10))
  }
  return out
})

const error = computed(() => (fetchError.value ? t('cra.planning_error') : ''))
</script>

<style scoped>
.muted {
  margin: 0;
  color: var(--kore-text-muted);
}
</style>
