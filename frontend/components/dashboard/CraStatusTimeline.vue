<script setup lang="ts">
import type { CraMonthItem } from '~/composables/useKpiMetrics'

const props = defineProps<{
  months: CraMonthItem[]
  loading?: boolean
  emptyLabel?: string
}>()

const { statusLabel } = useCraStatus()

function statusClass(status: string | null) {
  switch (status) {
    case 'Définitif':
      return 'cra-timeline__bar--validated'
    case 'ValidéSemaine':
      return 'cra-timeline__bar--submitted'
    case 'Brouillon':
      return 'cra-timeline__bar--draft'
    default:
      return 'cra-timeline__bar--empty'
  }
}

function statusText(status: string | null) {
  if (!status) return '—'
  return statusLabel(status)
}

const isEmpty = computed(() => !props.loading && props.months.every((m) => !m.status))
</script>

<template>
  <div class="cra-timeline" role="img" :aria-label="emptyLabel">
    <p v-if="loading" class="cra-timeline__state">{{ $t('common.loading') }}</p>
    <p v-else-if="isEmpty" class="cra-timeline__state">{{ emptyLabel }}</p>
    <div v-else class="cra-timeline__grid">
      <div v-for="month in months" :key="month.key" class="cra-timeline__col">
        <div
          class="cra-timeline__bar"
          :class="statusClass(month.status)"
          :title="`${month.label} — ${statusText(month.status)}`"
        />
        <span class="cra-timeline__label">{{ month.label }}</span>
      </div>
    </div>
    <ul v-if="!loading && !isEmpty" class="cra-timeline__legend">
      <li><span class="cra-timeline__dot cra-timeline__bar--validated" />{{ $t('cra.status_validated') }}</li>
      <li><span class="cra-timeline__dot cra-timeline__bar--submitted" />{{ $t('cra.status_submitted') }}</li>
      <li><span class="cra-timeline__dot cra-timeline__bar--draft" />{{ $t('cra.status_draft') }}</li>
      <li><span class="cra-timeline__dot cra-timeline__bar--empty" />{{ $t('dashboard.charts.cra_none') }}</li>
    </ul>
  </div>
</template>

<style scoped>
.cra-timeline__grid {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: var(--kore-space-sm);
  align-items: end;
  min-height: 8rem;
}

.cra-timeline__col {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--kore-space-xs);
  min-width: 0;
}

.cra-timeline__bar {
  width: 100%;
  max-width: 2.5rem;
  height: 5rem;
  border-radius: var(--kore-radius-md) var(--kore-radius-md) var(--kore-radius-sm) var(--kore-radius-sm);
  transition: height 0.2s ease;
}

.cra-timeline__bar--validated {
  background: var(--kore-success);
  height: 6rem;
}

.cra-timeline__bar--submitted {
  background: var(--kore-brand-gold);
  height: 4.5rem;
}

.cra-timeline__bar--draft {
  background: var(--kore-brand-blue-light);
  height: 3rem;
}

.cra-timeline__bar--empty {
  background: var(--kore-bg-subtle);
  border: 1px dashed var(--kore-border);
  height: 1.5rem;
}

.cra-timeline__label {
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
  text-transform: capitalize;
}

.cra-timeline__legend {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm) var(--kore-space-md);
  list-style: none;
  margin: var(--kore-space-md) 0 0;
  padding: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.cra-timeline__legend li {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}

.cra-timeline__dot {
  width: 0.625rem;
  height: 0.625rem;
  border-radius: var(--kore-radius-full);
  flex-shrink: 0;
}

.cra-timeline__state {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

@media (max-width: 640px) {
  .cra-timeline__grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    row-gap: var(--kore-space-md);
  }
}
</style>
