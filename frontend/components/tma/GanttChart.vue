<template>
  <div class="gantt">
    <div class="gantt__header">
      <div class="gantt__label-col">{{ labelHeader }}</div>
      <div class="gantt__timeline" role="presentation">
        <span
          v-for="tick in ticks"
          :key="tick.key"
          class="gantt__tick"
          :style="{ left: `${tick.pct}%` }"
        >
          {{ tick.label }}
        </span>
      </div>
    </div>

    <div v-if="!items.length" class="gantt__empty">
      {{ emptyTitle }}
    </div>

    <div
      v-for="row in computedRows"
      :key="row.id"
      class="gantt__row"
    >
      <div class="gantt__label-col">
        <span class="gantt__label" :title="row.label">{{ row.label }}</span>
        <AppBadge variant="neutral" class="gantt__badge">{{ row.statusLabel }}</AppBadge>
      </div>
      <div class="gantt__track" role="presentation">
        <div
          class="gantt__bar"
          :class="`gantt__bar--${row.statusTone}`"
          :style="row.barStyle"
          :title="row.tooltip"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { TmaGanttItem } from '~/composables/useTma'

const props = defineProps<{
  items: TmaGanttItem[]
  labelHeader: string
  emptyTitle: string
}>()

const { t, locale } = useI18n()

const range = computed(() => {
  if (!props.items.length) {
    const now = new Date()
    return { min: now, max: new Date(now.getTime() + 7 * 86400000) }
  }
  let min = props.items[0].start.getTime()
  let max = props.items[0].end.getTime()
  for (const item of props.items) {
    min = Math.min(min, item.start.getTime())
    max = Math.max(max, item.end.getTime())
  }
  const pad = Math.max(86400000, (max - min) * 0.05)
  return { min: new Date(min - pad), max: new Date(max + pad) }
})

const spanMs = computed(() => Math.max(range.value.max.getTime() - range.value.min.getTime(), 86400000))

const dateFmt = computed(() =>
  new Intl.DateTimeFormat(locale.value, { day: '2-digit', month: 'short' })
)

const ticks = computed(() => {
  const count = 5
  const out: Array<{ key: string; label: string; pct: number }> = []
  for (let i = 0; i < count; i++) {
    const pct = (i / (count - 1)) * 100
    const ts = range.value.min.getTime() + (spanMs.value * i) / (count - 1)
    out.push({
      key: String(i),
      label: dateFmt.value.format(new Date(ts)),
      pct
    })
  }
  return out
})

const statusLabel = (status: string) => {
  const key = `dashboard.charts.status.tma.${status}`
  const translated = t(key)
  return translated === key ? t('dashboard.charts.status.tma.unknown') : translated
}

const statusTone = (status: string): string => {
  switch (status) {
    case 'resolue':
      return 'success'
    case 'en_cours':
    case 'affectee':
      return 'blue'
    case 'en_attente_creation':
    case 'rework':
      return 'warn'
    case 'ouverte':
      return 'gold'
    default:
      return 'neutral'
  }
}

const computedRows = computed(() =>
  props.items.map((item) => {
    const left = ((item.start.getTime() - range.value.min.getTime()) / spanMs.value) * 100
    const width = Math.max(((item.end.getTime() - item.start.getTime()) / spanMs.value) * 100, 1.5)
    const label = statusLabel(item.status)
    return {
      id: item.id,
      label: item.label,
      statusLabel: label,
      statusTone: statusTone(item.status),
      barStyle: { left: `${left}%`, width: `${width}%` },
      tooltip: `${item.label} — ${label}`
    }
  })
)
</script>

<style scoped>
.gantt {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
  width: 100%;
}

.gantt__header,
.gantt__row {
  display: grid;
  grid-template-columns: minmax(8rem, 14rem) 1fr;
  gap: var(--kore-space-sm);
  align-items: center;
}

.gantt__label-col {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 0;
}

.gantt__label {
  font-size: var(--kore-text-small);
  font-weight: 600;
  color: var(--kore-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.gantt__badge {
  align-self: flex-start;
}

.gantt__timeline {
  position: relative;
  height: 1.5rem;
  border-bottom: 1px solid var(--kore-border);
}

.gantt__tick {
  position: absolute;
  bottom: 0;
  transform: translateX(-50%);
  font-size: 0.7rem;
  color: var(--kore-text-muted);
  white-space: nowrap;
}

.gantt__track {
  position: relative;
  height: 1.75rem;
  background: var(--kore-bg-muted);
  border-radius: var(--kore-radius-sm);
  overflow: hidden;
}

.gantt__bar {
  position: absolute;
  top: 0.25rem;
  height: 1.25rem;
  border-radius: var(--kore-radius-sm);
  min-width: 4px;
}

.gantt__bar--success { background: var(--kore-status-success); }
.gantt__bar--blue { background: var(--kore-status-info, var(--kore-brand-blue)); }
.gantt__bar--warn { background: var(--kore-status-warn); }
.gantt__bar--gold { background: var(--kore-brand-gold); }
.gantt__bar--neutral { background: var(--kore-text-muted); }

.gantt__empty {
  padding: var(--kore-space-lg);
  text-align: center;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

@media (max-width: 768px) {
  .gantt__header,
  .gantt__row {
    grid-template-columns: 1fr;
  }

  .gantt__timeline {
    display: none;
  }

  .gantt__header .gantt__label-col {
    display: none;
  }

  .gantt__track {
    height: 0.5rem;
  }

  .gantt__bar {
    top: 0.125rem;
    height: 0.25rem;
  }
}
</style>
