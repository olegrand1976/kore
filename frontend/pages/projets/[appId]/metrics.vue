<template>
  <div>
    <AppPageHeader :title="$t('project.metrics_title')" :subtitle="appLabel">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/projets')">{{ $t('project.back') }}</AppButton>
      </template>
    </AppPageHeader>

    <div class="metrics-grid">
      <AppCard padding="lg">
        <h2 class="metrics-card__title">{{ $t('project.velocity_title') }}</h2>
        <p v-if="pending" class="muted">{{ $t('project.loading') }}</p>
        <template v-else>
          <p class="metrics-highlight">{{ velocity?.averageVelocity?.toFixed(1) ?? '0' }} {{ $t('project.story_points_abbr') }}</p>
          <ul v-if="velocity?.sprints?.length" class="metrics-list">
            <li v-for="s in velocity.sprints" :key="s.sprintName">
              {{ s.sprintName }} — {{ s.closedPoints }} {{ $t('project.story_points_abbr') }}
            </li>
          </ul>
          <p v-else class="muted">{{ $t('project.velocity_empty') }}</p>
        </template>
      </AppCard>

      <AppCard v-if="burndown" padding="lg" class="metrics-burndown">
        <h2 class="metrics-card__title">{{ $t('project.burndown_title') }}</h2>
        <p class="muted">{{ $t('project.burndown_planned', { points: burndown.plannedPoints }) }}</p>
        <div class="burndown-chart" role="img" :aria-label="$t('project.burndown_title')">
          <svg viewBox="0 0 400 160" class="burndown-chart__svg">
            <polyline
              class="burndown-chart__ideal"
              :points="idealPointsAttr"
            />
            <polyline
              class="burndown-chart__actual"
              :points="actualPointsAttr"
            />
          </svg>
          <div class="burndown-chart__legend">
            <span class="burndown-chart__key burndown-chart__key--ideal">{{ $t('project.burndown_ideal') }}</span>
            <span class="burndown-chart__key burndown-chart__key--actual">{{ $t('project.burndown_actual') }}</span>
          </div>
        </div>
      </AppCard>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const route = useRoute()
const appId = computed(() => String(route.params.appId ?? ''))
const { get, pickAppLabel } = useApplications()
const { getVelocity, getBurndown, listSprints, pickSprintId } = useProject()

const appLabel = ref('')
const pending = ref(true)
const velocity = ref<{ averageVelocity: number; sprints: { sprintName: string; closedPoints: number }[] } | null>(null)
const burndown = ref<{ plannedPoints: number; points: { date: string; remainingPoints: number; idealPoints: number }[] } | null>(null)

const chartMax = computed(() => {
  const pts = burndown.value?.points ?? []
  if (!pts.length) return 1
  return Math.max(burndown.value?.plannedPoints ?? 1, ...pts.map((p) => Math.max(p.remainingPoints, p.idealPoints)), 1)
})

function seriesToAttr(key: 'remainingPoints' | 'idealPoints') {
  const pts = burndown.value?.points ?? []
  if (!pts.length) return ''
  const max = chartMax.value
  const w = 380
  const h = 140
  return pts
    .map((p, i) => {
      const x = 10 + (i / Math.max(pts.length - 1, 1)) * w
      const y = 10 + h - (p[key] / max) * h
      return `${x},${y}`
    })
    .join(' ')
}

const idealPointsAttr = computed(() => seriesToAttr('idealPoints'))
const actualPointsAttr = computed(() => seriesToAttr('remainingPoints'))

onMounted(async () => {
  try {
    const app = await get(appId.value)
    appLabel.value = pickAppLabel(app)
    velocity.value = (await getVelocity(appId.value)) ?? null
    const sprints = await listSprints(appId.value)
    const active = sprints.find((s) => (s.status ?? s.Status) === 'active') ?? sprints[0]
    if (active) {
      burndown.value = (await getBurndown(appId.value, pickSprintId(active))) ?? null
    }
  } finally {
    pending.value = false
  }
})
</script>

<style scoped>
.metrics-grid {
  display: grid;
  gap: var(--kore-space-md);
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
}

.metrics-card__title {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-h3);
}

.metrics-highlight {
  font-size: var(--kore-text-h2);
  font-weight: 600;
  margin: 0 0 var(--kore-space-md);
}

.metrics-list {
  margin: 0;
  padding-left: var(--kore-space-md);
}

.metrics-burndown {
  grid-column: 1 / -1;
}

.burndown-chart__svg {
  width: 100%;
  max-width: 640px;
  height: auto;
  display: block;
}

.burndown-chart__ideal {
  fill: none;
  stroke: var(--kore-text-muted);
  stroke-width: 2;
  stroke-dasharray: 4 4;
}

.burndown-chart__actual {
  fill: none;
  stroke: var(--kore-brand-gold);
  stroke-width: 2;
}

.burndown-chart__legend {
  display: flex;
  gap: var(--kore-space-md);
  margin-top: var(--kore-space-sm);
  font-size: var(--kore-text-small);
}

.burndown-chart__key::before {
  content: '';
  display: inline-block;
  width: 1rem;
  height: 2px;
  margin-right: var(--kore-space-xs);
  vertical-align: middle;
}

.burndown-chart__key--ideal::before {
  background: var(--kore-text-muted);
}

.burndown-chart__key--actual::before {
  background: var(--kore-brand-gold);
}
</style>
