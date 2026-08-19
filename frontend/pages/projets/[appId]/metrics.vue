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

      <AppCard v-if="burndown" padding="lg">
        <h2 class="metrics-card__title">{{ $t('project.burndown_title') }}</h2>
        <p class="muted">{{ $t('project.burndown_planned', { points: burndown.plannedPoints }) }}</p>
        <ul class="metrics-list">
          <li v-for="p in burndown.points" :key="p.date">
            {{ p.date }} — {{ p.remainingPoints }} / {{ p.idealPoints }} {{ $t('project.story_points_abbr') }}
          </li>
        </ul>
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
</style>
