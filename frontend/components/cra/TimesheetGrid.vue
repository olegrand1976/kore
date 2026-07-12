<template>
  <AppCard padding="lg">
    <div class="grid-header">
      <h3 class="section-title">{{ $t('cra.weeks_title') }}</h3>
      <div class="week-tabs" role="tablist">
        <button
          v-for="n in weekNumbers"
          :key="n"
          type="button"
          role="tab"
          class="week-tab"
          :class="{ 'week-tab--active': n === activeWeek }"
          :aria-selected="n === activeWeek"
          @click="activeWeek = n"
        >
          {{ $t('cra.week_n', { n }) }}
        </button>
      </div>
    </div>

    <WeekEditor
      :week-number="activeWeek"
      :week="currentWeek"
      :month="month"
      :disabled="!canEdit"
      :saving="saving"
      @save="onSave"
      @submit="onSubmit"
    />
  </AppCard>
</template>

<script setup lang="ts">
import type { CraLine, CraWeek } from '~/stores/cra'

const props = defineProps<{
  weeks: CraWeek[]
  month: string
  canEdit: boolean
  saving?: boolean
}>()

const emit = defineEmits<{
  save: [weekNumber: number, lines: CraLine[]]
  submit: [weekNumber: number]
}>()

const weekNumbers = [1, 2, 3, 4, 5]
const activeWeek = ref(1)

const currentWeek = computed(() =>
  props.weeks.find((w) => w.weekNumber === activeWeek.value)
)

const onSave = (lines: CraLine[]) => emit('save', activeWeek.value, lines)
const onSubmit = () => emit('submit', activeWeek.value)
</script>

<style scoped>
.grid-header {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  align-items: center;
  gap: var(--kore-space-md);
  margin-bottom: var(--kore-space-lg);
}

.section-title {
  margin: 0;
  font-size: var(--kore-text-h3);
}

.week-tabs {
  display: flex;
  gap: var(--kore-space-xs);
  flex-wrap: wrap;
}

.week-tab {
  padding: 0.375rem 0.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  color: var(--kore-text-muted);
  cursor: pointer;
  font-size: var(--kore-text-small);
}

.week-tab--active {
  border-color: var(--kore-brand-gold);
  color: var(--kore-brand-gold);
  background: rgba(201, 162, 39, 0.1);
}
</style>
