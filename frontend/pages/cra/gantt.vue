<template>
  <div>
    <AppPageHeader :title="$t('cra.gantt_title')" :subtitle="$t('cra.gantt_subtitle')">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/cra/planning')">
          {{ $t('cra.planning_link') }}
        </AppButton>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/cra')">
          <AppIcon name="arrow_back" /> {{ $t('cra.back') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard padding="lg">
      <p v-if="pending" class="muted">{{ $t('cra.loading') }}</p>
      <p v-else-if="error" class="flash flash--error" role="alert">{{ error }}</p>
      <CraGanttChart
        v-else
        :items="items"
        :label-header="$t('cra.gantt_col_mission')"
        :empty-title="$t('cra.gantt_empty')"
      />
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const { fetchGantt } = useReporting()
const { t } = useI18n()

const { data, pending, error: fetchError } = await useAsyncData('cra-gantt', () => fetchGantt({ window: '60' }))

const items = computed(() => data.value ?? [])
const error = computed(() => (fetchError.value ? t('cra.gantt_error') : ''))
</script>

<style scoped>
.muted {
  margin: 0;
  color: var(--kore-text-muted);
}
</style>
