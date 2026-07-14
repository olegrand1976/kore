<template>
  <AppCard padding="lg">
    <div class="grid-header">
      <h3 class="section-title">{{ $t('cra.weeks_title') }}</h3>
      <div class="week-tabs" role="tablist">
        <button
          v-for="tab in weekTabs"
          :key="tab.weekNumber"
          type="button"
          role="tab"
          class="week-tab"
          :class="{ 'week-tab--active': tab.weekNumber === activeWeek }"
          :aria-selected="tab.weekNumber === activeWeek"
          @click="activeWeek = tab.weekNumber"
        >
          <span class="week-tab__label">{{ weekTabLabel(tab) }}</span>
          <AppIcon v-if="isWeekSubmitted(tab.weekNumber)" name="check_circle" class="week-tab__check" />
        </button>
      </div>
    </div>

    <WeekMatrix
      :week-number="activeWeek"
      :week="currentWeek"
      :month="month"
      :week-start-day="weekStartDay"
      :day-capacity-minutes="dayCapacityMinutes"
      :week-submit-policy="weekSubmitPolicy"
      :week-label="activeTabLabel"
      :disabled="!canEdit"
      :saving="saving"
      :missions="missions"
      :task-types="taskTypes"
      @save="onSave"
      @submit="onSubmit"
    />
  </AppCard>
</template>

<script setup lang="ts">
import type { CraLine, CraWeek } from '~/stores/cra'
import type { MissionSummary } from '~/composables/useCraSourceLabels'
import { computeMonthWeeks } from '~/composables/useWeekCalendar'

const props = defineProps<{
  weeks: CraWeek[]
  month: string
  weekStartDay: number
  dayCapacityMinutes?: number
  weekSubmitPolicy?: 'block' | 'warn' | 'none'
  canEdit: boolean
  saving?: boolean
  missions?: MissionSummary[]
  taskTypes?: string[]
}>()

const emit = defineEmits<{
  save: [weekNumber: number, lines: CraLine[]]
  submit: [weekNumber: number]
}>()

const { t, locale } = useI18n()

const weekTabs = computed(() => computeMonthWeeks(props.month, props.weekStartDay))
const activeWeek = ref(weekTabs.value[0]?.weekNumber ?? 1)

watch(weekTabs, (tabs) => {
  if (!tabs.some((tab) => tab.weekNumber === activeWeek.value)) {
    activeWeek.value = tabs[0]?.weekNumber ?? 1
  }
})

const currentWeek = computed(() => props.weeks.find((w) => w.weekNumber === activeWeek.value))

const activeTab = computed(() => weekTabs.value.find((t) => t.weekNumber === activeWeek.value))

const activeTabLabel = computed(() => {
  const tab = activeTab.value
  if (!tab) return t('cra.week_n', { n: activeWeek.value })
  const fmt = (raw: string) =>
    new Date(`${raw}T12:00:00`).toLocaleDateString(locale.value === 'en' ? 'en-US' : 'fr-FR', {
      day: 'numeric',
      month: 'short'
    })
  return `${t('cra.week_n', { n: tab.weekNumber })} (${fmt(tab.start)} – ${fmt(tab.end)})`
})

const weekTabLabel = (tab: { weekNumber: number; start: string; end: string }) => {
  const fmt = (raw: string) =>
    new Date(`${raw}T12:00:00`).toLocaleDateString(locale.value === 'en' ? 'en-US' : 'fr-FR', {
      day: 'numeric',
      month: 'short'
    })
  return `${t('cra.week_n', { n: tab.weekNumber })} (${fmt(tab.start)} – ${fmt(tab.end)})`
}

const isWeekSubmitted = (weekNumber: number) => Boolean(props.weeks.find((w) => w.weekNumber === weekNumber)?.submittedAt)

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
  flex-wrap: nowrap;
  overflow-x: auto;
  scroll-snap-type: x mandatory;
  max-width: 100%;
  padding-bottom: 0.25rem;
}

.week-tab {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.375rem 0.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  color: var(--kore-text-muted);
  cursor: pointer;
  font-size: var(--kore-text-small);
  white-space: nowrap;
  scroll-snap-align: start;
  flex: 0 0 auto;
}

.week-tab--active {
  border-color: var(--kore-brand-gold);
  color: var(--kore-brand-gold);
  background: rgba(201, 162, 39, 0.1);
}

.week-tab__check {
  font-size: 1rem;
  color: var(--kore-success);
}
</style>
