<template>
  <div>
    <AppPageHeader :title="$t('project.title')" :subtitle="$t('project.subtitle')" />

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('project.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="!apps.length" padding="lg">
      <AppEmptyState icon="view_kanban" :title="$t('project.empty')" :description="$t('project.empty_hint')" />
    </AppCard>

    <div v-else class="projets-grid">
      <AppCard v-for="app in apps" :key="app.id" padding="lg" class="projets-card">
        <h2 class="projets-card__title">{{ app.label }}</h2>
        <AppBadge variant="blue">{{ profileLabel(app.profile) }}</AppBadge>
        <div class="projets-card__links">
          <NuxtLink :to="`/projets/${app.id}/backlog`" class="projets-link">{{ $t('project.nav_backlog') }}</NuxtLink>
          <NuxtLink v-if="app.profile === 'agile_scrum'" :to="`/projets/${app.id}/sprints`" class="projets-link">
            {{ $t('project.nav_sprints') }}
          </NuxtLink>
          <NuxtLink :to="`/projets/${app.id}/epics`" class="projets-link">{{ $t('project.nav_epics') }}</NuxtLink>
          <NuxtLink v-if="app.profile === 'agile_kanban'" :to="`/projets/${app.id}/board`" class="projets-link">
            {{ $t('project.nav_board') }}
          </NuxtLink>
          <NuxtLink v-if="app.profile === 'agile_kanban'" :to="`/projets/${app.id}/kanban-config`" class="projets-link">
            {{ $t('project.nav_kanban_config') }}
          </NuxtLink>
          <NuxtLink :to="`/projets/${app.id}/metrics`" class="projets-link">{{ $t('project.nav_metrics') }}</NuxtLink>
        </div>
      </AppCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { MethodologyProfile } from '~/composables/useMethodologyTerms'

definePageMeta({ layout: 'default' })

const { t } = useI18n()
const { listAgileApplications } = useProject()
const { pickAppId, pickAppLabel, pickAppMethodologyProfile } = useApplications()

const pending = ref(true)
const apps = ref<{ id: string; label: string; profile: MethodologyProfile }[]>([])

const profileLabel = (profile: MethodologyProfile) => {
  switch (profile) {
    case 'agile_scrum':
      return t('project.profile_scrum')
    case 'agile_kanban':
      return t('project.profile_kanban')
    default:
      return t('project.profile_psa')
  }
}

onMounted(async () => {
  try {
    const raw = await listAgileApplications()
    apps.value = raw.map((app) => ({
      id: pickAppId(app),
      label: pickAppLabel(app),
      profile: pickAppMethodologyProfile(app)
    }))
  } finally {
    pending.value = false
  }
})
</script>

<style scoped>
.projets-grid {
  display: grid;
  gap: var(--kore-space-md);
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
}

.projets-card__title {
  margin: 0 0 var(--kore-space-sm);
  font-size: var(--kore-text-h3);
}

.projets-card__links {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-sm);
  margin-top: var(--kore-space-md);
}

.projets-link {
  color: var(--kore-link);
  text-decoration: none;
}

@media (max-width: 768px) {
  .projets-grid {
    grid-template-columns: 1fr;
  }
}
</style>
