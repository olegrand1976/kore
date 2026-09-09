<template>
  <AppCard padding="lg">
    <div class="grid-header">
      <h3 class="section-title">{{ $t('cra.weeks_title') }}</h3>
      <div ref="weekTabsEl" class="week-tabs" role="tablist">
        <button
          v-for="tab in weekTabs"
          :key="tab.weekNumber"
          type="button"
          role="tab"
          class="week-tab"
          :class="{ 'week-tab--active': tab.weekNumber === activeWeek }"
          :aria-selected="tab.weekNumber === activeWeek"
          :data-week="tab.weekNumber"
          @click="lockActiveWeek(tab.weekNumber)"
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
      :planned-week-minutes="plannedWeekMinutes"
      :missions="missions"
      :task-types="taskTypes"
      :work-ref-options="workRefOptions"
      :work-ref-label-for="workRefLabelFor"
      @save="onSave"
      @submit="onSubmit"
    />
  </AppCard>
</template>

<script setup lang="ts">
import type { CraLine, CraWeek } from '~/stores/cra'
import type { MissionSummary } from '~/composables/useCraSourceLabels'
import { computeMonthWeeks, initialActiveWeekNumber } from '~/composables/useWeekCalendar'

import type { CraWorkRefOption } from '~/composables/useCraWorkRefs'

const props = defineProps<{
  weeks: CraWeek[]
  month: string
  weekStartDay: number
  dayCapacityMinutes?: number
  weekSubmitPolicy?: 'block' | 'warn' | 'none'
  canEdit: boolean
  saving?: boolean
  plannedWeekMinutes?: number | null
  missions?: MissionSummary[]
  taskTypes?: string[]
  workRefOptions?: CraWorkRefOption[]
  workRefLabelFor?: (type: string, id: string) => string
}>()

const emit = defineEmits<{
  save: [weekNumber: number, lines: CraLine[]]
  submit: [weekNumber: number]
}>()

/** null = auto (current week if month is current); set = locked selection (survives save / remount). */
const activeWeekModel = defineModel<number | null>('activeWeek', { default: null })

const { t, locale } = useI18n()

const weekTabsEl = ref<HTMLElement | null>(null)
const weekTabs = computed(() => computeMonthWeeks(props.month, props.weekStartDay))

const activeWeek = computed({
  get: () => activeWeekModel.value ?? initialActiveWeekNumber(props.month, props.weekStartDay),
  set: (weekNumber: number) => {
    activeWeekModel.value = weekNumber
  }
})

const lockActiveWeek = (weekNumber: number) => {
  activeWeekModel.value = weekNumber
}

watch(() => props.weekStartDay, () => {
  if (activeWeekModel.value == null) return
  if (!weekTabs.value.some((tab) => tab.weekNumber === activeWeekModel.value)) {
    activeWeekModel.value = weekTabs.value[0]?.weekNumber ?? 1
  }
})

/** Scroll only the tab strip — avoid page jump from Element.scrollIntoView. */
const scrollActiveTabIntoView = () => {
  const root = weekTabsEl.value
  if (!root || typeof root.querySelector !== 'function') return
  const active = root.querySelector<HTMLElement>(`.week-tab[data-week="${activeWeek.value}"]`)
  if (!active) return
  const tabLeft = active.offsetLeft
  const tabRight = tabLeft + active.offsetWidth
  const viewLeft = root.scrollLeft
  const viewRight = viewLeft + root.clientWidth
  if (tabLeft < viewLeft) {
    root.scrollTo({ left: tabLeft, behavior: 'smooth' })
  } else if (tabRight > viewRight) {
    root.scrollTo({ left: tabRight - root.clientWidth, behavior: 'smooth' })
  }
}

onMounted(() => {
  nextTick(scrollActiveTabIntoView)
})

watch(activeWeek, () => {
  nextTick(scrollActiveTabIntoView)
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

const onSave = (lines: CraLine[]) => {
  const weekNumber = activeWeek.value
  lockActiveWeek(weekNumber)
  emit('save', weekNumber, lines)
}

const onSubmit = () => {
  const weekNumber = activeWeek.value
  lockActiveWeek(weekNumber)
  emit('submit', weekNumber)
}
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
