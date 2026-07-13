<script setup lang="ts">
import type { ChartBarItem } from '~/composables/useKpiMetrics'

const props = withDefaults(
  defineProps<{
    bars: ChartBarItem[]
    maxValue?: number
    loading?: boolean
    emptyLabel?: string
    valueSuffix?: string
  }>(),
  { loading: false, valueSuffix: '' }
)

const max = computed(() => {
  if (props.maxValue != null && props.maxValue > 0) return props.maxValue
  const peak = Math.max(...props.bars.map((b) => b.value), 0)
  return peak > 0 ? peak : 1
})

const isEmpty = computed(() => !props.loading && props.bars.every((b) => b.value === 0))

function barWidth(value: number) {
  return `${Math.max(4, Math.round((value / max.value) * 100))}%`
}

function toneClass(tone?: ChartBarItem['tone']) {
  switch (tone) {
    case 'gold':
      return 'app-bar-chart__fill--gold'
    case 'blue':
      return 'app-bar-chart__fill--blue'
    case 'success':
      return 'app-bar-chart__fill--success'
    case 'warn':
      return 'app-bar-chart__fill--warn'
    case 'muted':
      return 'app-bar-chart__fill--muted'
    case undefined:
      return 'app-bar-chart__fill--gold'
    default: {
      const _exhaustive: never = tone
      return _exhaustive
    }
  }
}
</script>

<template>
  <div class="app-bar-chart" role="img" :aria-label="emptyLabel">
    <p v-if="loading" class="app-bar-chart__state">{{ $t('common.loading') }}</p>
    <p v-else-if="isEmpty" class="app-bar-chart__state">{{ emptyLabel }}</p>
    <ul v-else class="app-bar-chart__list">
      <li v-for="bar in bars" :key="bar.key" class="app-bar-chart__row">
        <span class="app-bar-chart__label" :title="bar.label">{{ bar.label }}</span>
        <div class="app-bar-chart__track">
          <div
            class="app-bar-chart__fill"
            :class="toneClass(bar.tone)"
            :style="{ width: barWidth(bar.value) }"
          />
        </div>
        <span class="app-bar-chart__value">{{ bar.value }}{{ valueSuffix }}</span>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.app-bar-chart__list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--kore-space-sm);
}

.app-bar-chart__row {
  display: grid;
  grid-template-columns: minmax(4.5rem, 7rem) 1fr minmax(2.5rem, auto);
  align-items: center;
  gap: var(--kore-space-sm);
}

.app-bar-chart__label {
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.app-bar-chart__track {
  height: 0.625rem;
  background: var(--kore-bg-subtle);
  border-radius: var(--kore-radius-full);
  overflow: hidden;
}

.app-bar-chart__fill {
  height: 100%;
  border-radius: var(--kore-radius-full);
  transition: width 0.25s ease;
}

.app-bar-chart__fill--gold {
  background: var(--kore-brand-gold);
}

.app-bar-chart__fill--blue {
  background: var(--kore-brand-blue-light);
}

.app-bar-chart__fill--success {
  background: var(--kore-success);
}

.app-bar-chart__fill--warn {
  background: var(--kore-error);
}

.app-bar-chart__fill--muted {
  background: var(--kore-border);
}

.app-bar-chart__value {
  font-size: var(--kore-text-caption);
  font-weight: 600;
  color: var(--kore-text);
  text-align: right;
  white-space: nowrap;
}

.app-bar-chart__state {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

@media (max-width: 640px) {
  .app-bar-chart__row {
    grid-template-columns: 1fr auto;
    grid-template-rows: auto auto;
  }

  .app-bar-chart__label {
    grid-column: 1;
  }

  .app-bar-chart__value {
    grid-column: 2;
    grid-row: 1;
  }

  .app-bar-chart__track {
    grid-column: 1 / -1;
  }
}
</style>
