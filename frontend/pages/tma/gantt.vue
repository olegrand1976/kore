<template>
  <div>
    <AppPageHeader :title="$t('tma.gantt_title')" :subtitle="$t('tma.gantt_subtitle')">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/tma')">
          {{ $t('tma.back') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard padding="lg">
      <p v-if="pending" class="gantt-loading">{{ $t('tma.loading') }}</p>
      <GanttChart
        v-else
        :items="ganttItems"
        :label-header="$t('tma.gantt_col_task')"
        :empty-title="$t('tma.gantt_empty')"
      />
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const { list, toGanttItems } = useTma()

const { data, pending } = await useAsyncData('tma-gantt', () => list())

const ganttItems = computed(() => toGanttItems(data.value ?? []))
</script>

<style scoped>
.gantt-loading {
  margin: 0;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}
</style>
