<template>
  <div v-if="visible" class="cra-sticky-pill" role="status" :aria-label="ariaLabel">
    <span class="cra-sticky-pill__hours">{{ hoursLabel }}</span>
    <span class="cra-sticky-pill__sep">·</span>
    <span class="cra-sticky-pill__pct">{{ progress }}%</span>
  </div>
</template>

<script setup lang="ts">
import { minutesToHoursLabel } from '~/composables/useWeekCalendar'

const props = defineProps<{
  totalMinutes: number
  capacityMinutes: number
  visible?: boolean
}>()

const { t } = useI18n()

const progress = computed(() => {
  if (props.capacityMinutes <= 0) return 0
  return Math.min(100, Math.round((props.totalMinutes / props.capacityMinutes) * 100))
})

const hoursLabel = computed(() => {
  const total = minutesToHoursLabel(props.totalMinutes)
  const cap = props.capacityMinutes ? minutesToHoursLabel(props.capacityMinutes) : null
  return cap ? `${total}h / ${cap}h` : `${total}h`
})

const ariaLabel = computed(() =>
  t('cra.sticky_totals_aria', { hours: hoursLabel.value, pct: progress.value })
)
</script>

<style scoped>
.cra-sticky-pill {
  position: fixed;
  top: calc(3.5rem + env(safe-area-inset-top, 0px));
  left: 50%;
  transform: translateX(-50%);
  z-index: 3;
  display: flex;
  align-items: center;
  gap: var(--kore-space-xs);
  padding: var(--kore-space-xs) var(--kore-space-md);
  border-radius: 999px;
  border: 1px solid var(--kore-border);
  background: var(--kore-bg-elevated);
  box-shadow: var(--kore-shadow-sm);
  font-size: var(--kore-text-small);
  font-weight: 600;
}

.cra-sticky-pill__sep {
  color: var(--kore-text-muted);
}

.cra-sticky-pill__pct {
  color: var(--kore-brand-gold);
}

@media (min-width: 900px) {
  .cra-sticky-pill {
    display: none;
  }
}
</style>
